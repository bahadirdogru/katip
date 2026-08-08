package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"katip/internal/diff"
	"katip/internal/hardware"
	"katip/internal/kitap"
	"katip/internal/llm"
	"katip/internal/tarz"
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
	ModelPath        string `json:"modelPath"`
	ServerBinary     string `json:"serverBinary"`
	ServerHost       string `json:"serverHost"`
	ServerPort       int    `json:"serverPort"`
	CtxSize          int    `json:"ctxSize"`
	Threads          int    `json:"threads"`
	SystemPrompt     string `json:"systemPrompt"`
	GPULayers        int    `json:"gpuLayers"`
	Backend          string `json:"backend"`
	AutoStartLLM     bool   `json:"autoStartLLM"`
	HuggingFaceToken string `json:"huggingFaceToken"`
}

func defaultConfig() *AppConfig {
	profile := hardware.GetProfile()
	return &AppConfig{
		ServerHost:   "127.0.0.1",
		ServerPort:   8089,
		CtxSize:      4096,
		Threads:      hardware.OptimalThreads(profile),
		SystemPrompt: defaultSystemPrompt,
		GPULayers:    -1,
		Backend:      "auto",
		AutoStartLLM: true,
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
	ChangeType  string     `json:"changeType,omitempty"`
	RuleName    string     `json:"ruleName,omitempty"`
}

type DiffItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *KatipService) Greet(name string) string {
	return "Merhaba " + name + "! Katip hazır."
}

func (s *KatipService) ImproveParagraph(paragraphID string, text string, plotSummary string, mode string) (*DiffResult, error) {
	return s.improveText(paragraphID, text, plotSummary, llm.ImprovementMode(mode))
}

func (s *KatipService) ImproveSelection(selectionID string, text string, plotSummary string, mode string) (*DiffResult, error) {
	return s.improveText(selectionID, text, plotSummary, llm.ImprovementMode(mode))
}

func (s *KatipService) improveText(id string, text string, plotSummary string, mode llm.ImprovementMode) (*DiffResult, error) {
	if mode == "" {
		mode = llm.ModeFix
	}

	improved, err := s.llmClient.ImproveWithOptions(text, plotSummary, mode)
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
		summary = formatChangeSummary(diffItems, changeCount)
	}

	return &DiffResult{
		ParagraphID: id,
		Summary:     summary,
		Original:    text,
		Improved:    improved,
		Diffs:       diffItems,
		ChangeType:  inferChangeType(diffItems),
	}, nil
}

func (s *KatipService) CancelImprovement() {
	s.llmClient.Cancel()
}

func (s *KatipService) GetImproveStreamProgress() map[string]interface{} {
	text, streaming := s.llmClient.GetStreamProgress()
	return map[string]interface{}{
		"text":      text,
		"streaming": streaming,
	}
}

func (s *KatipService) CheckConsistency(text string, plotSummary string) (string, error) {
	return s.llmClient.CheckConsistency(text, plotSummary)
}

func (s *KatipService) GetTarzProfiles() ([]string, error) {
	tarz.EnsureDefaultProfile()
	return tarz.ListProfiles()
}

func (s *KatipService) CheckTarzStyle(profileName string, text string) ([]tarz.StyleViolation, error) {
	tarz.EnsureDefaultProfile()
	if profileName == "" {
		profileName = "varsayilan"
	}
	profile, err := tarz.LoadProfile(profileName)
	if err != nil {
		return nil, err
	}
	return tarz.CheckText(profile, text), nil
}

func (s *KatipService) GeneratePlotSummary(fullText string) (string, error) {
	return s.llmClient.GenerateSummary(fullText)
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
	threads := s.config.Threads
	if threads <= 0 {
		threads = hardware.OptimalThreads(hardware.GetProfile())
	}
	gpuLayers := s.config.GPULayers
	profile := hardware.GetProfile()
	backend := llm.ResolveBackend(s.config.Backend)
	if backend == "cpu" || !profile.HasGPUAcceleration {
		gpuLayers = 0
	} else if gpuLayers == 0 {
		gpuLayers = -1
	}
	return s.llmManager.Start(llm.ServerConfig{
		BinaryPath: s.config.ServerBinary,
		ModelPath:  s.config.ModelPath,
		Host:       s.config.ServerHost,
		Port:       s.config.ServerPort,
		CtxSize:    s.config.CtxSize,
		Threads:    threads,
		GPULayers:  gpuLayers,
	})
}

func (s *KatipService) StopLLMServer() error {
	return s.llmManager.Stop()
}

func (s *KatipService) CheckSetupStatus() map[string]interface{} {
	llamaInstalled := llm.IsLlamaServerInstalled()
	llamaPath := llm.GetLlamaServerPath()
	zipPath := llm.FindExistingZip()

	profile := hardware.GetProfile()
	recommendedID := llm.GetRecommendedModelID()
	recommended := llm.FindModelByID(recommendedID)
	if recommended == nil {
		recommended = llm.GetDefaultModel()
	}

	modelInstalled := false
	modelPath := ""
	var modelPartialBytes int64
	recommendedModelID := recommendedID
	recommendedModelName := ""
	recommendedModelSize := ""

	if recommended != nil {
		recommendedModelID = recommended.ID
		recommendedModelName = recommended.Name
		recommendedModelSize = recommended.SizeLabel
		if llm.IsModelDownloaded(recommended.Filename) {
			modelInstalled = true
			modelPath = llm.GetModelPath(recommended.Filename)
		} else if _, partSize := llm.FindModelPartFile(recommended.Filename); partSize > 0 {
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
	if s.config.Threads <= 0 {
		s.config.Threads = hardware.OptimalThreads(profile)
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
		"recommendedModelID":   recommendedModelID,
		"defaultModelName":     recommendedModelName,
		"defaultModelSize":     recommendedModelSize,
		"defaultModelID":       recommendedModelID,
		"hardwareSummary":      formatHardwareSummary(profile),
		"recommendedBackend":   profile.RecommendedBackend,
		"hasGPU":               profile.HasGPUAcceleration,
	}

	return result
}

func formatHardwareSummary(p hardware.Profile) string {
	totalGB := float64(p.TotalRAMBytes) / (1024 * 1024 * 1024)
	if p.HasGPUAcceleration && len(p.GPUs) > 0 {
		return fmt.Sprintf("%.0f GB RAM, %s (%d MB VRAM)", totalGB, p.GPUs[0].Name, p.GPUs[0].VRAMMB)
	}
	if p.RecommendedBackend == "metal" {
		return fmt.Sprintf("%.0f GB RAM, Apple Silicon (Metal)", totalGB)
	}
	return fmt.Sprintf("%.0f GB RAM, CPU modu", totalGB)
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
	return llm.EnrichCatalogWithFit(hardware.GetProfile())
}

func (s *KatipService) GetHardwareProfile() hardware.Profile {
	return hardware.GetProfile()
}

func (s *KatipService) RefreshHardwareProfile() hardware.Profile {
	return hardware.RefreshProfile()
}

func (s *KatipService) GetSystemMonitor() map[string]interface{} {
	profile := hardware.GetProfile()
	return map[string]interface{}{
		"totalRAMBytes":     profile.TotalRAMBytes,
		"availableRAMBytes": profile.AvailableRAMBytes,
		"usedRAMBytes":      profile.UsedRAMBytes,
		"cpuCores":          profile.CPUCores,
		"gpus":              profile.GPUs,
		"llmRunning":        s.llmManager.IsRunning(),
		"llmHealthy":        s.llmClient.IsHealthy(),
	}
}

func (s *KatipService) SelectModel(modelID string) error {
	path, err := llm.GetModelPathForID(modelID)
	if err != nil {
		return err
	}
	wasRunning := s.llmManager.IsRunning()
	if wasRunning {
		_ = s.llmManager.Stop()
	}
	s.config.ModelPath = path
	if err := s.saveConfig(); err != nil {
		return err
	}
	if wasRunning {
		return s.StartLLMServer()
	}
	return nil
}

func (s *KatipService) DeleteModel(modelID string) error {
	if s.config.ModelPath != "" {
		m := llm.FindModelByID(modelID)
		if m != nil && s.config.ModelPath == llm.GetModelPath(m.Filename) {
			if s.llmManager.IsRunning() {
				_ = s.llmManager.Stop()
			}
			s.config.ModelPath = ""
			_ = s.saveConfig()
		}
	}
	return llm.DeleteModel(modelID)
}

func (s *KatipService) GetDiskUsage() (llm.DiskUsage, error) {
	return llm.GetDiskUsage()
}

func (s *KatipService) ListHuggingFaceGGUF(repoID string) ([]llm.HFRepoFile, error) {
	return llm.ListHuggingFaceGGUF(repoID, s.config.HuggingFaceToken)
}

func (s *KatipService) DownloadHuggingFaceModel(repoID string, filename string) error {
	if s.modelDownloading {
		return fmt.Errorf("model indirmesi zaten devam ediyor")
	}
	s.modelDownloading = true
	s.modelDownloadProgress = &llm.DownloadProgress{Status: "Başlatılıyor...", Percent: 0}

	go func() {
		defer func() { s.modelDownloading = false }()
		err := llm.DownloadHuggingFaceFile(repoID, filename, s.config.HuggingFaceToken, func(p llm.DownloadProgress) {
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
		s.config.ModelPath = llm.GetModelPath(filename)
		s.saveConfig()
	}()
	return nil
}

func (s *KatipService) DownloadLlamaServerForBackend(backend string) error {
	if s.downloading {
		return fmt.Errorf("indirme zaten devam ediyor")
	}
	if backend == "" || backend == "auto" {
		backend = llm.ResolveBackend(s.config.Backend)
	}
	s.downloading = true
	s.downloadProgress = &llm.DownloadProgress{Status: "Başlatılıyor...", Percent: 0}

	go func() {
		defer func() { s.downloading = false }()
		err := llm.DownloadLlamaServerWithBackend(backend, func(p llm.DownloadProgress) {
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
		if backend != "cpu" {
			s.config.Backend = backend
		}
		s.saveConfig()
	}()
	return nil
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

func formatChangeSummary(items []DiffItem, count int) string {
	if count == 1 {
		for _, d := range items {
			if d.Type == "delete" {
				for _, d2 := range items {
					if d2.Type == "insert" {
						preview := d.Text
						if len(preview) > 15 {
							preview = preview[:15] + "..."
						}
						ins := d2.Text
						if len(ins) > 15 {
							ins = ins[:15] + "..."
						}
						return fmt.Sprintf(`"%s" → "%s"`, preview, ins)
					}
				}
			}
		}
		return "1 düzeltme önerildi."
	}
	return fmt.Sprintf("%d düzeltme önerildi.", count)
}

func inferChangeType(items []DiffItem) string {
	for _, d := range items {
		if d.Type == "delete" || d.Type == "insert" {
			return "grammar"
		}
	}
	return "style"
}

func (s *KatipService) SaveKitap(filePath string, doc kitap.Document) error {
	return kitap.Write(&doc, filePath)
}

func (s *KatipService) LoadKitap(filePath string) (*kitap.Document, error) {
	return kitap.Read(filePath)
}
