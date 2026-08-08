package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DiskUsage struct {
	TotalBytes int64 `json:"totalBytes"`
	ModelCount int   `json:"modelCount"`
}

func GetDiskUsage() (DiskUsage, error) {
	dir := GetModelDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return DiskUsage{}, nil
		}
		return DiskUsage{}, err
	}

	var total int64
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".gguf") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		total += info.Size()
		count++
	}
	return DiskUsage{TotalBytes: total, ModelCount: count}, nil
}

func DeleteModel(modelID string) error {
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

	path := GetModelPath(model.Filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("model silinemedi: %w", err)
	}
	_ = os.Remove(path + ".part")
	return nil
}

func FindModelByID(modelID string) *ModelInfo {
	for i, m := range ModelCatalog {
		if m.ID == modelID {
			return &ModelCatalog[i]
		}
	}
	return nil
}

func GetModelPathForID(modelID string) (string, error) {
	m := FindModelByID(modelID)
	if m == nil {
		return "", fmt.Errorf("model bulunamadı: %s", modelID)
	}
	if !IsModelDownloaded(m.Filename) {
		return "", fmt.Errorf("model indirilmemiş: %s", modelID)
	}
	return GetModelPath(m.Filename), nil
}

type HFRepoFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
}

func ListHuggingFaceGGUF(repoID, token string) ([]HFRepoFile, error) {
	repoID = strings.TrimSpace(repoID)
	repoID = strings.Trim(repoID, "/")
	if repoID == "" {
		return nil, fmt.Errorf("HuggingFace repo ID boş olamaz")
	}

	apiURL := fmt.Sprintf("https://huggingface.co/api/models/%s/tree/main", repoID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Katip")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HuggingFace API hatası: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("erişim reddedildi — gated model için HF token gerekli")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HuggingFace API HTTP %d: %s", resp.StatusCode, string(body))
	}

	var entries []struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("yanıt okunamadı: %w", err)
	}

	var files []HFRepoFile
	for _, e := range entries {
		if e.Type != "file" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(e.Path), ".gguf") {
			continue
		}
		filename := filepath.Base(e.Path)
		url := fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repoID, e.Path)
		files = append(files, HFRepoFile{
			Filename: filename,
			Size:     e.Size,
			URL:      url,
		})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("repoda GGUF dosyası bulunamadı")
	}
	return files, nil
}

func DownloadHuggingFaceFile(repoID, filename, token string, progressCb func(DownloadProgress)) error {
	repoID = strings.TrimSpace(strings.Trim(repoID, "/"))
	files, err := ListHuggingFaceGGUF(repoID, token)
	if err != nil {
		return err
	}

	var target *HFRepoFile
	for i, f := range files {
		if f.Filename == filename {
			target = &files[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("dosya bulunamadı: %s", filename)
	}

	report := func(p DownloadProgress) {
		if progressCb != nil {
			progressCb(p)
		}
	}

	destDir := GetModelDir()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, target.Filename)
	partPath := destPath + ".part"

	var resumeFrom int64
	if info, err := os.Stat(partPath); err == nil {
		resumeFrom = info.Size()
	}

	report(DownloadProgress{
		Status:  fmt.Sprintf("%s indiriliyor...", target.Filename),
		Percent: 0,
		Total:   target.Size,
	})

	req, err := http.NewRequest("GET", target.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Katip")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("indirme hatası: HTTP %d", resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		resumeFrom = 0
	}

	outFile, err := os.OpenFile(partPath, flags, 0644)
	if err != nil {
		return err
	}

	downloaded := resumeFrom
	total := target.Size
	if total <= 0 {
		total = resp.ContentLength + resumeFrom
	}

	buf := make([]byte, 64*1024)
	lastPct := -1
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := outFile.Write(buf[:n]); wErr != nil {
				outFile.Close()
				return wErr
			}
			downloaded += int64(n)
			if total > 0 {
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct != lastPct {
					lastPct = pct
					report(DownloadProgress{
						Status:     "İndiriliyor...",
						Percent:    pct,
						Downloaded: downloaded,
						Total:      total,
					})
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			outFile.Close()
			return readErr
		}
	}
	outFile.Close()

	if err := os.Rename(partPath, destPath); err != nil {
		return err
	}

	report(DownloadProgress{Status: "Tamamlandı!", Percent: 100, Downloaded: downloaded, Total: total})
	return nil
}
