package llm

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubAPILatest = "https://api.github.com/repos/ggml-org/llama.cpp/releases/latest"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type DownloadProgress struct {
	Status     string `json:"status"`
	Percent    int    `json:"percent"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Error      string `json:"error,omitempty"`
}

func GetLlamaServerDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "Katip", "llama-server")
}

func GetLlamaServerPath() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return filepath.Join(GetLlamaServerDir(), "llama-server"+ext)
}

func IsLlamaServerInstalled() bool {
	path := GetLlamaServerPath()
	_, err := os.Stat(path)
	return err == nil
}

func FindExistingZip() string {
	dir := GetLlamaServerDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".zip") {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func findAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch {
	case os == "windows" && arch == "amd64":
		return "win-cpu-x64"
	case os == "windows" && arch == "arm64":
		return "win-cpu-arm64"
	case os == "darwin" && arch == "arm64":
		return "mac-arm64"
	case os == "darwin" && arch == "amd64":
		return "mac-x64"
	case os == "linux" && arch == "amd64":
		return "ubuntu-x64"
	case os == "linux" && arch == "arm64":
		return "ubuntu-arm64"
	default:
		return "win-cpu-x64"
	}
}

func DownloadLlamaServer(progressCb func(DownloadProgress)) error {
	report := func(p DownloadProgress) {
		if progressCb != nil {
			progressCb(p)
		}
	}

	report(DownloadProgress{Status: "Sürüm bilgisi alınıyor...", Percent: 0})

	client := &http.Client{}
	req, _ := http.NewRequest("GET", githubAPILatest, nil)
	req.Header.Set("User-Agent", "Katip")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub API'ye bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("sürüm bilgisi okunamadı: %w", err)
	}

	pattern := findAssetName()
	var targetAsset *githubAsset
	for i, asset := range release.Assets {
		if strings.Contains(asset.Name, pattern) &&
			!strings.Contains(asset.Name, "cuda") &&
			!strings.Contains(asset.Name, "vulkan") &&
			!strings.Contains(asset.Name, "sycl") &&
			!strings.Contains(asset.Name, "hip") {
			targetAsset = &release.Assets[i]
			break
		}
	}
	if targetAsset == nil {
		return fmt.Errorf("platformunuz için uygun binary bulunamadı (%s)", pattern)
	}

	slog.Info("llama-server indiriliyor", "version", release.TagName, "asset", targetAsset.Name, "size_mb", targetAsset.Size/1024/1024)
	report(DownloadProgress{Status: fmt.Sprintf("%s indiriliyor (%d MB)...", release.TagName, targetAsset.Size/1024/1024), Percent: 5, Total: targetAsset.Size})

	destDir := GetLlamaServerDir()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("dizin oluşturulamadı: %w", err)
	}

	zipPath := filepath.Join(destDir, targetAsset.Name)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı: %w", err)
	}

	dlResp, err := http.Get(targetAsset.BrowserDownloadURL)
	if err != nil {
		zipFile.Close()
		return fmt.Errorf("indirme başlatılamadı: %w", err)
	}
	defer dlResp.Body.Close()

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := dlResp.Body.Read(buf)
		if n > 0 {
			zipFile.Write(buf[:n])
			downloaded += int64(n)
			pct := int(float64(downloaded) / float64(targetAsset.Size) * 80)
			if pct > 80 {
				pct = 80
			}
			report(DownloadProgress{Status: "İndiriliyor...", Percent: 5 + pct, Downloaded: downloaded, Total: targetAsset.Size})
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			zipFile.Close()
			os.Remove(zipPath)
			return fmt.Errorf("indirme hatası: %w", readErr)
		}
	}
	zipFile.Close()

	report(DownloadProgress{Status: "Arşiv açılıyor...", Percent: 88})

	if err := ExtractLlamaServerZip(zipPath, destDir); err != nil {
		return fmt.Errorf("arşiv açılamadı: %w", err)
	}

	report(DownloadProgress{Status: "Tamamlandı!", Percent: 100})
	slog.Info("llama-server kuruldu", "path", GetLlamaServerPath())

	return nil
}

func ExtractLlamaServerZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	foundServer := false
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		name := filepath.Base(f.Name)
		ext := strings.ToLower(filepath.Ext(name))

		keep := ext == ".exe" || ext == ".dll" || ext == ".so" || ext == ".dylib" || ext == ".metal" || name == "LICENSE"
		if !keep {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outPath := filepath.Join(destDir, name)
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}

		if name == "llama-server.exe" || name == "llama-server" {
			foundServer = true
		}
		slog.Info("arşivden çıkarıldı", "file", name)
	}

	if !foundServer {
		return fmt.Errorf("arşivde llama-server bulunamadı")
	}
	return nil
}
