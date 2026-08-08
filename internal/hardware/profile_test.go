package hardware

import "testing"

func TestComputeFitStatus(t *testing.T) {
	profile := Profile{
		TotalRAMBytes:     16 * 1024 * 1024 * 1024,
		AvailableRAMBytes: 12 * 1024 * 1024 * 1024,
		GPUs:              []GPUInfo{{VRAMMB: 8192, Backend: "cuda"}},
		HasGPUAcceleration: true,
		RecommendedBackend: "cuda",
	}

	if got := ComputeFitStatus(2*1024*1024*1024, profile); got != "fits" {
		t.Fatalf("expected fits, got %s", got)
	}
	if got := ComputeFitStatus(20*1024*1024*1024, profile); got != "wont_fit" {
		t.Fatalf("expected wont_fit, got %s", got)
	}
}

func TestRecommendedModelID(t *testing.T) {
	low := Profile{TotalRAMBytes: 4 * 1024 * 1024 * 1024}
	if id := RecommendedModelID(low); id != "bitnet-2b-4t" {
		t.Fatalf("expected bitnet, got %s", id)
	}
	high := Profile{TotalRAMBytes: 32 * 1024 * 1024 * 1024}
	if id := RecommendedModelID(high); id != "turkcell-7b-q4km" {
		t.Fatalf("expected turkcell, got %s", id)
	}
}

func TestGetProfile(t *testing.T) {
	p := GetProfile()
	if p.CPUCores <= 0 {
		t.Fatal("cpu cores should be positive")
	}
	if p.TotalRAMBytes == 0 {
		t.Fatal("total ram should be positive")
	}
}
