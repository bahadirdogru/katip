package llm

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SizeLabel   string `json:"sizeLabel"`
	SizeBytes   int64  `json:"sizeBytes"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	Language    string `json:"language"`
	MinRAM      string `json:"minRAM"`
	IsDefault   bool   `json:"isDefault"`
}

var ModelCatalog = []ModelInfo{
	{
		ID:          "turkcell-7b-q4km",
		Name:        "Turkcell-LLM-7b-v1",
		Description: "Mistral 7B tabanlı, 5 milyar Türkçe token ile eğitilmiş. En iyi Türkçe kalitesi.",
		SizeLabel:   "~4.5 GB",
		SizeBytes:   4_787_937_536,
		Filename:    "Turkcell-LLM-7b-v1.Q4_K_M.gguf",
		URL:         "https://huggingface.co/QuantFactory/Turkcell-LLM-7b-v1-GGUF/resolve/main/Turkcell-LLM-7b-v1.Q4_K_M.gguf",
		Language:    "Türkçe",
		MinRAM:      "8 GB",
		IsDefault:   true,
	},
	{
		ID:          "openr1-qwen-7b-tr-q4km",
		Name:        "OpenR1-Qwen-7B-Turkish",
		Description: "Qwen2.5 tabanlı, Türkçe reasoning/düşünme yeteneği. Dolphin-R1 dataset ile fine-tune.",
		SizeLabel:   "~4.5 GB",
		SizeBytes:   4_683_218_944,
		Filename:    "OpenR1-Qwen-7B-Turkish.Q4_K_M.gguf",
		URL:         "https://huggingface.co/mradermacher/OpenR1-Qwen-7B-Turkish-GGUF/resolve/main/OpenR1-Qwen-7B-Turkish.Q4_K_M.gguf",
		Language:    "Türkçe",
		MinRAM:      "8 GB",
	},
	{
		ID:          "qwen25-3b-q4km",
		Name:        "Qwen2.5-3B-Instruct",
		Description: "Hafif ve hızlı çok dilli model. Sınırlı kaynaklı sistemler için ideal.",
		SizeLabel:   "~2.0 GB",
		SizeBytes:   2_058_000_000,
		Filename:    "qwen2.5-3b-instruct-q4_k_m.gguf",
		URL:         "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf",
		Language:    "Çok dilli",
		MinRAM:      "4 GB",
	},
	{
		ID:          "bitnet-2b-4t",
		Name:        "BitNet b1.58-2B-4T",
		Description: "Microsoft'un 1.58-bit ultra hızlı modeli. Çok düşük kaynak kullanımı, Türkçe desteği zayıf.",
		SizeLabel:   "~1.1 GB",
		SizeBytes:   1_187_801_280,
		Filename:    "ggml-model-i2_s.gguf",
		URL:         "https://huggingface.co/microsoft/bitnet-b1.58-2B-4T-gguf/resolve/main/ggml-model-i2_s.gguf",
		Language:    "İngilizce",
		MinRAM:      "2 GB",
	},
}

func GetModelDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "Katip", "models")
}

func GetModelPath(filename string) string {
	return filepath.Join(GetModelDir(), filename)
}

func IsModelDownloaded(filename string) bool {
	path := GetModelPath(filename)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

func GetDefaultModel() *ModelInfo {
	for i, m := range ModelCatalog {
		if m.IsDefault {
			return &ModelCatalog[i]
		}
	}
	return nil
}

func FindModelPartFile(filename string) (string, int64) {
	partPath := GetModelPath(filename) + ".part"
	if info, err := os.Stat(partPath); err == nil && info.Size() > 0 {
		return partPath, info.Size()
	}
	return "", 0
}

func GetInstalledModels() []string {
	var installed []string
	for _, m := range ModelCatalog {
		if IsModelDownloaded(m.Filename) {
			installed = append(installed, m.ID)
		}
	}
	return installed
}

func DownloadModel(modelID string, progressCb func(DownloadProgress)) error {
	var model *ModelInfo
	for i, m := range ModelCatalog {
		if m.ID == modelID {
			model = &ModelCatalog[i]
			break
		}
	}
	if model == nil {
		return fmt.Errorf("model bulunamadı: %s", modelID)
	}

	report := func(p DownloadProgress) {
		if progressCb != nil {
			progressCb(p)
		}
	}

	destDir := GetModelDir()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("dizin oluşturulamadı: %w", err)
	}

	destPath := filepath.Join(destDir, model.Filename)

	var resumeFrom int64
	if info, err := os.Stat(destPath + ".part"); err == nil {
		resumeFrom = info.Size()
	}

	slog.Info("model indiriliyor", "id", modelID, "name", model.Name, "size", model.SizeLabel)
	report(DownloadProgress{
		Status:  fmt.Sprintf("%s indiriliyor (%s)...", model.Name, model.SizeLabel),
		Percent: 0,
		Total:   model.SizeBytes,
	})

	req, err := http.NewRequest("GET", model.URL, nil)
	if err != nil {
		return fmt.Errorf("istek oluşturulamadı: %w", err)
	}
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
		report(DownloadProgress{
			Status:     fmt.Sprintf("Kaldığı yerden devam ediliyor (%d MB)...", resumeFrom/1024/1024),
			Percent:    int(float64(resumeFrom) / float64(model.SizeBytes) * 100),
			Downloaded: resumeFrom,
			Total:      model.SizeBytes,
		})
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("indirme başlatılamadı: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("sunucu hatası: HTTP %d", resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		resumeFrom = 0
	}

	partPath := destPath + ".part"
	outFile, err := os.OpenFile(partPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı: %w", err)
	}

	downloaded := resumeFrom
	buf := make([]byte, 64*1024)
	lastReportedPct := -1

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := outFile.Write(buf[:n]); wErr != nil {
				outFile.Close()
				return fmt.Errorf("yazma hatası: %w", wErr)
			}
			downloaded += int64(n)
			pct := int(float64(downloaded) / float64(model.SizeBytes) * 100)
			if pct > 99 {
				pct = 99
			}
			if pct != lastReportedPct {
				lastReportedPct = pct
				report(DownloadProgress{
					Status:     "İndiriliyor...",
					Percent:    pct,
					Downloaded: downloaded,
					Total:      model.SizeBytes,
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			outFile.Close()
			return fmt.Errorf("indirme hatası: %w", readErr)
		}
	}
	outFile.Close()

	if err := os.Rename(partPath, destPath); err != nil {
		return fmt.Errorf("dosya taşınamadı: %w", err)
	}

	report(DownloadProgress{Status: "Tamamlandı!", Percent: 100, Downloaded: downloaded, Total: model.SizeBytes})
	slog.Info("model indirildi", "id", modelID, "path", destPath)

	return nil
}
