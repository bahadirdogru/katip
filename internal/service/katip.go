package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"katip/internal/diff"
	"katip/internal/llm"
)

type KatipService struct {
	llmClient             *llm.Client
	llmManager            *llm.Manager
	diffEngine            *diff.Engine
	config                *AppConfig
	configPath            string
	downloadProgress      *llm.DownloadProgress
	downloading           bool
	modelDownloadProgress *llm.DownloadProgress
	modelDownloading      bool
}

type AppConfig struct {
	ModelPath     string `json:"modelPath"`
	ServerBinary  string `json:"serverBinary"`
	ServerHost    string `json:"serverHost"`
	ServerPort    int    `json:"serverPort"`
	CtxSize       int    `json:"ctxSize"`
	Threads       int    `json:"threads"`
	SystemPrompt  string `json:"systemPrompt"`
}

func defaultConfig() *AppConfig {
	return &AppConfig{
		ServerHost:   "127.0.0.1",
		ServerPort:   8089,
		CtxSize:      4096,
		Threads:      4,
		SystemPrompt: defaultSystemPrompt,
	}
}

const defaultSystemPrompt = `Sen bir metin düzeltme motorusun. Sohbet YAPMA. Soru SORMA. Açıklama YAPMA.

GİRDİ: <DÜZELT> etiketi arasında Türkçe metin alacaksın.
ÇIKTI: Yalnızca düzeltilmiş metni döndür. Başka hiçbir şey yazma.

KURALLAR:
- Yazım ve noktalama hatalarını düzelt.
- Anlatım bozukluklarını düzelt.
- Cümle akışını iyileştir.
- Metnin anlamını, uzunluğunu ve yapısını KORU.
- Yeni cümle, bilgi veya yorum EKLEME.
- Tırnak işareti, önek, etiket veya açıklama EKLEME.
- Metin doğruysa aynen döndür.
- Yanıtın SADECE düzeltilmiş metin olmalı, başka HİÇBİR ŞEY olmamalı.`

func NewKatipService() *KatipService {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "Katip", "config.json")

	cfg := loadConfig(configPath)
	endpoint := fmt.Sprintf("http://%s:%d", cfg.ServerHost, cfg.ServerPort)

	client := llm.NewClient(endpoint)
	if cfg.SystemPrompt != "" {
		client.SetSystemPrompt(cfg.SystemPrompt)
	}

	return &KatipService{
		llmClient:  client,
		llmManager: llm.NewManager(),
		diffEngine: diff.NewEngine(),
		config:     cfg,
		configPath: configPath,
	}
}

func loadConfig(path string) *AppConfig {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, cfg)
	return cfg
}

func (s *KatipService) saveConfig() error {
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

type DiffResult struct {
	ParagraphID string     `json:"paragraphId"`
	Summary     string     `json:"summary"`
	Original    string     `json:"original"`
	Improved    string     `json:"improved"`
	Diffs       []DiffItem `json:"diffs"`
}

type DiffItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *KatipService) Greet(name string) string {
	return "Merhaba " + name + "! Katip hazır."
}

func (s *KatipService) ImproveParagraph(paragraphID string, text string) (*DiffResult, error) {
	improved, err := s.llmClient.Improve(text)
	if err != nil {
		return nil, err
	}

	diffs := s.diffEngine.ComputeWordDiff(text, improved)

	diffItems := make([]DiffItem, len(diffs))
	changeCount := 0
	for i, d := range diffs {
		diffItems[i] = DiffItem{
			Type: d.Type,
			Text: d.Text,
		}
		if d.Type != "equal" {
			changeCount++
		}
	}

	summary := ""
	if changeCount == 0 {
		summary = "Değişiklik önerilmedi."
	} else {
		summary = formatChangeSummary(changeCount)
	}

	return &DiffResult{
		ParagraphID: paragraphID,
		Summary:     summary,
		Original:    text,
		Improved:    improved,
		Diffs:       diffItems,
	}, nil
}

func (s *KatipService) GetLLMStatus() map[string]interface{} {
	return map[string]interface{}{
		"running":   s.llmManager.IsRunning(),
		"healthy":   s.llmClient.IsHealthy(),
		"endpoint":  s.llmClient.Endpoint(),
		"modelPath": s.config.ModelPath,
		"lastError": s.llmManager.LastError(),
	}
}

func (s *KatipService) GetServerLog() string {
	return s.llmManager.Log()
}

func (s *KatipService) GetConfig() *AppConfig {
	return s.config
}

func (s *KatipService) UpdateConfig(cfg AppConfig) error {
	s.config = &cfg
	endpoint := fmt.Sprintf("http://%s:%d", cfg.ServerHost, cfg.ServerPort)
	s.llmClient.SetEndpoint(endpoint)
	if cfg.SystemPrompt != "" {
		s.llmClient.SetSystemPrompt(cfg.SystemPrompt)
	}
	return s.saveConfig()
}

func (s *KatipService) StartLLMServer() error {
	if s.config.ServerBinary == "" {
		return fmt.Errorf("llama-server binary yolu ayarlanmamış")
	}
	if s.config.ModelPath == "" {
		return fmt.Errorf("model dosyası yolu ayarlanmamış")
	}
	return s.llmManager.Start(llm.ServerConfig{
		BinaryPath: s.config.ServerBinary,
		ModelPath:  s.config.ModelPath,
		Host:       s.config.ServerHost,
		Port:       s.config.ServerPort,
		CtxSize:    s.config.CtxSize,
		Threads:    s.config.Threads,
	})
}

func (s *KatipService) StopLLMServer() error {
	return s.llmManager.Stop()
}

func (s *KatipService) CheckSetupStatus() map[string]interface{} {
	llamaInstalled := llm.IsLlamaServerInstalled()
	llamaPath := llm.GetLlamaServerPath()
	zipPath := llm.FindExistingZip()

	defModel := llm.GetDefaultModel()
	modelInstalled := false
	modelPath := ""
	var modelPartialBytes int64

	if defModel != nil {
		if llm.IsModelDownloaded(defModel.Filename) {
			modelInstalled = true
			modelPath = llm.GetModelPath(defModel.Filename)
		} else if _, partSize := llm.FindModelPartFile(defModel.Filename); partSize > 0 {
			modelPartialBytes = partSize
		}
	}

	configUpdated := false
	if llamaInstalled && s.config.ServerBinary == "" {
		s.config.ServerBinary = llamaPath
		configUpdated = true
	}
	if modelInstalled && s.config.ModelPath == "" {
		s.config.ModelPath = modelPath
		configUpdated = true
	}
	if configUpdated {
		s.saveConfig()
	}

	status := "ready"
	if !llamaInstalled && zipPath != "" {
		status = "zip_found"
	} else if !llamaInstalled {
		status = "llama_missing"
	} else if !modelInstalled && modelPartialBytes > 0 {
		status = "model_partial"
	} else if !modelInstalled {
		status = "model_missing"
	}

	result := map[string]interface{}{
		"status":            status,
		"llamaInstalled":    llamaInstalled,
		"llamaPath":         llamaPath,
		"zipExists":         zipPath != "",
		"modelInstalled":    modelInstalled,
		"modelPath":         modelPath,
		"modelPartialBytes": modelPartialBytes,
	}

	if defModel != nil {
		result["defaultModelName"] = defModel.Name
		result["defaultModelSize"] = defModel.SizeLabel
		result["defaultModelID"] = defModel.ID
	}

	return result
}

func (s *KatipService) CheckLlamaServer() map[string]interface{} {
	installed := llm.IsLlamaServerInstalled()
	path := llm.GetLlamaServerPath()
	zipExists := llm.FindExistingZip() != ""
	return map[string]interface{}{
		"installed": installed,
		"path":      path,
		"zipExists": zipExists,
	}
}

func (s *KatipService) DownloadLlamaServer() error {
	if s.downloading {
		return fmt.Errorf("indirme zaten devam ediyor")
	}
	s.downloading = true
	s.downloadProgress = &llm.DownloadProgress{Status: "Başlatılıyor...", Percent: 0}

	go func() {
		defer func() { s.downloading = false }()
		err := llm.DownloadLlamaServer(func(p llm.DownloadProgress) {
			s.downloadProgress = &p
		})
		if err != nil {
			s.downloadProgress = &llm.DownloadProgress{
				Status:  "Hata: " + err.Error(),
				Percent: -1,
				Error:   err.Error(),
			}
			return
		}
		path := llm.GetLlamaServerPath()
		s.config.ServerBinary = path
		s.saveConfig()
	}()
	return nil
}

func (s *KatipService) ReextractLlamaServer() error {
	zipPath := llm.FindExistingZip()
	if zipPath == "" {
		return fmt.Errorf("indirilen zip dosyası bulunamadı, önce indirmeniz gerekiyor")
	}
	destDir := llm.GetLlamaServerDir()
	if err := llm.ExtractLlamaServerZip(zipPath, destDir); err != nil {
		return fmt.Errorf("arşiv açılamadı: %w", err)
	}
	path := llm.GetLlamaServerPath()
	s.config.ServerBinary = path
	s.saveConfig()
	return nil
}

func (s *KatipService) GetDownloadProgress() *llm.DownloadProgress {
	if s.downloadProgress == nil {
		return &llm.DownloadProgress{Status: "", Percent: 0}
	}
	return s.downloadProgress
}

func (s *KatipService) GetModelCatalog() []llm.ModelInfo {
	result := make([]llm.ModelInfo, len(llm.ModelCatalog))
	copy(result, llm.ModelCatalog)
	return result
}

func (s *KatipService) GetInstalledModels() []string {
	return llm.GetInstalledModels()
}

func (s *KatipService) DownloadModel(modelID string) error {
	if s.modelDownloading {
		return fmt.Errorf("model indirmesi zaten devam ediyor")
	}
	s.modelDownloading = true
	s.modelDownloadProgress = &llm.DownloadProgress{Status: "Başlatılıyor...", Percent: 0}

	go func() {
		defer func() { s.modelDownloading = false }()
		err := llm.DownloadModel(modelID, func(p llm.DownloadProgress) {
			s.modelDownloadProgress = &p
		})
		if err != nil {
			s.modelDownloadProgress = &llm.DownloadProgress{
				Status:  "Hata: " + err.Error(),
				Percent: -1,
				Error:   err.Error(),
			}
			return
		}
		for _, m := range llm.ModelCatalog {
			if m.ID == modelID {
				s.config.ModelPath = llm.GetModelPath(m.Filename)
				s.saveConfig()
				break
			}
		}
	}()
	return nil
}

func (s *KatipService) GetModelDownloadProgress() *llm.DownloadProgress {
	if s.modelDownloadProgress == nil {
		return &llm.DownloadProgress{Status: "", Percent: 0}
	}
	return s.modelDownloadProgress
}

func formatChangeSummary(count int) string {
	if count == 1 {
		return "1 düzeltme önerildi."
	}
	return fmt.Sprintf("%d düzeltme önerildi.", count)
}
