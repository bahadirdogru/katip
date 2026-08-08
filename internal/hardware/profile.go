package hardware

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GPUInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	VRAMMB  uint64 `json:"vramMB"`
	Backend string `json:"backend"`
}

type Profile struct {
	TotalRAMBytes      uint64    `json:"totalRAMBytes"`
	AvailableRAMBytes  uint64    `json:"availableRAMBytes"`
	UsedRAMBytes       uint64    `json:"usedRAMBytes"`
	CPUCores           int       `json:"cpuCores"`
	GPUs               []GPUInfo `json:"gpus"`
	RecommendedBackend string    `json:"recommendedBackend"`
	HasGPUAcceleration bool      `json:"hasGPUAcceleration"`
	Platform           string    `json:"platform"`
	Arch               string    `json:"arch"`
}

var (
	cachedProfile *Profile
	cacheMu       sync.RWMutex
	cacheTime     time.Time
	cacheTTL      = 30 * time.Second
)

func GetProfile() Profile {
	cacheMu.RLock()
	if cachedProfile != nil && time.Since(cacheTime) < cacheTTL {
		p := *cachedProfile
		cacheMu.RUnlock()
		return p
	}
	cacheMu.RUnlock()

	p := detectProfile()

	cacheMu.Lock()
	cachedProfile = &p
	cacheTime = time.Now()
	cacheMu.Unlock()

	return p
}

func RefreshProfile() Profile {
	cacheMu.Lock()
	cachedProfile = nil
	cacheMu.Unlock()
	return GetProfile()
}

func detectProfile() Profile {
	total, avail, used := readMemory()
	gpus := detectGPUs()

	backend := "cpu"
	hasGPU := false

	if len(gpus) > 0 {
		hasGPU = true
		backend = gpus[0].Backend
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		hasGPU = true
		backend = "metal"
		gpus = []GPUInfo{{
			ID:      "0",
			Name:    "Apple Silicon (Metal)",
			VRAMMB:  total / 1024 / 1024 / 2,
			Backend: "metal",
		}}
	}

	return Profile{
		TotalRAMBytes:      total,
		AvailableRAMBytes:  avail,
		UsedRAMBytes:       used,
		CPUCores:           runtime.NumCPU(),
		GPUs:               gpus,
		RecommendedBackend: backend,
		HasGPUAcceleration: hasGPU,
		Platform:           runtime.GOOS,
		Arch:               runtime.GOARCH,
	}
}

func readMemory() (total, avail, used uint64) {
	switch runtime.GOOS {
	case "linux":
		return readLinuxMemory()
	case "windows":
		return readWindowsMemory()
	case "darwin":
		return readDarwinMemory()
	default:
		return 8 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024
	}
}

func readLinuxMemory() (total, avail, used uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 8 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024
	}
	var memTotal, memAvail uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024
		switch fields[0] {
		case "MemTotal:":
			memTotal = val
		case "MemAvailable:":
			memAvail = val
		}
	}
	if memTotal == 0 {
		memTotal = 8 * 1024 * 1024 * 1024
	}
	if memAvail == 0 {
		memAvail = memTotal / 2
	}
	return memTotal, memAvail, memTotal - memAvail
}

func readWindowsMemory() (total, avail, used uint64) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize, FreePhysicalMemory | ConvertTo-Json)")
	out, err := cmd.Output()
	if err != nil {
		return 8 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024
	}
	text := string(out)
	totalKB := extractJSONNumber(text, "TotalVisibleMemorySize")
	freeKB := extractJSONNumber(text, "FreePhysicalMemory")
	if totalKB == 0 {
		return 8 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024
	}
	total = totalKB * 1024
	avail = freeKB * 1024
	used = total - avail
	return total, avail, used
}

func extractJSONNumber(json, key string) uint64 {
	idx := strings.Index(json, key)
	if idx < 0 {
		return 0
	}
	rest := json[idx+len(key):]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return 0
	}
	rest = strings.TrimSpace(rest[colon+1:])
	end := 0
	for end < len(rest) && (rest[end] >= '0' && rest[end] <= '9') {
		end++
	}
	if end == 0 {
		return 0
	}
	v, _ := strconv.ParseUint(rest[:end], 10, 64)
	return v
}

func readDarwinMemory() (total, avail, used uint64) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	out, err := cmd.Output()
	if err != nil {
		return 8 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024, 4 * 1024 * 1024 * 1024
	}
	total, _ = strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	cmd2 := exec.Command("vm_stat")
	out2, err := cmd2.Output()
	if err != nil {
		avail = total / 2
		used = total - avail
		return total, avail, used
	}
	pageSize := uint64(4096)
	var free, inactive, speculative uint64
	scanner := bufio.NewScanner(bytes.NewReader(out2))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		valStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "."))
		val, _ := strconv.ParseUint(valStr, 10, 64)
		switch strings.TrimSpace(parts[0]) {
		case "Pages free":
			free = val
		case "Pages inactive":
			inactive = val
		case "Pages speculative":
			speculative = val
		}
	}
	avail = (free + inactive + speculative) * pageSize
	if avail > total {
		avail = total / 2
	}
	used = total - avail
	return total, avail, used
}

func detectGPUs() []GPUInfo {
	nvidia := detectNvidiaGPUs()
	if len(nvidia) > 0 {
		return nvidia
	}
	return nil
}

func detectNvidiaGPUs() []GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,memory.total",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 3 {
			continue
		}
		vram, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		gpus = append(gpus, GPUInfo{
			ID:      strings.TrimSpace(parts[0]),
			Name:    strings.TrimSpace(parts[1]),
			VRAMMB:  vram,
			Backend: "cuda",
		})
	}
	return gpus
}

// FitStatus: fits | may_be_slow | wont_fit
func ComputeFitStatus(modelSizeBytes int64, profile Profile) string {
	if modelSizeBytes <= 0 {
		return "fits"
	}

	ramBudget := int64(float64(profile.TotalRAMBytes) * 0.80)
	size := modelSizeBytes

	if size > ramBudget {
		return "wont_fit"
	}

	if profile.HasGPUAcceleration && len(profile.GPUs) > 0 {
		maxVRAM := profile.GPUs[0].VRAMMB * 1024 * 1024
		for _, g := range profile.GPUs[1:] {
			v := g.VRAMMB * 1024 * 1024
			if v > maxVRAM {
				maxVRAM = v
			}
		}
		vramBudget := int64(float64(maxVRAM) * 0.80)
		if size > vramBudget && size <= ramBudget {
			return "may_be_slow"
		}
		if size > ramBudget {
			return "wont_fit"
		}
		return "fits"
	}

	slowThreshold := int64(float64(profile.TotalRAMBytes) * 0.50)
	if size > slowThreshold {
		return "may_be_slow"
	}
	return "fits"
}

func RecommendedModelID(profile Profile) string {
	totalGB := float64(profile.TotalRAMBytes) / (1024 * 1024 * 1024)
	switch {
	case totalGB < 6:
		return "bitnet-2b-4t"
	case totalGB < 10:
		return "qwen25-3b-q4km"
	default:
		return "turkcell-7b-q4km"
	}
}

func OptimalThreads(profile Profile) int {
	cores := profile.CPUCores
	if cores <= 1 {
		return 1
	}
	if cores > 8 {
		return 8
	}
	return cores
}

func HighestVRAMMB(profile Profile) uint64 {
	var max uint64
	for _, g := range profile.GPUs {
		if g.VRAMMB > max {
			max = g.VRAMMB
		}
	}
	return max
}
