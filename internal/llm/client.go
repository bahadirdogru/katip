package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ImprovementMode string

const (
	ModeFix     ImprovementMode = "fix"
	ModeShorten ImprovementMode = "shorten"
	ModeFlow    ImprovementMode = "flow"
	ModeFormal  ImprovementMode = "formal"
)

type Client struct {
	endpoint     string
	httpClient   *http.Client
	systemPrompt string
	cancelMu     sync.Mutex
	cancelled    bool
	streamMu     sync.Mutex
	streamText   string
	streaming    bool
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		systemPrompt: defaultSystemPrompt,
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) SetEndpoint(endpoint string) {
	c.endpoint = endpoint
}

func (c *Client) SetSystemPrompt(prompt string) {
	c.systemPrompt = prompt
}

func (c *Client) Cancel() {
	c.cancelMu.Lock()
	c.cancelled = true
	c.cancelMu.Unlock()
}

func (c *Client) resetCancel() {
	c.cancelMu.Lock()
	c.cancelled = false
	c.cancelMu.Unlock()
}

func (c *Client) isCancelled() bool {
	c.cancelMu.Lock()
	defer c.cancelMu.Unlock()
	return c.cancelled
}

func (c *Client) GetStreamProgress() (string, bool) {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()
	return c.streamText, c.streaming
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

var modePrompts = map[ImprovementMode]string{
	ModeFix:     "",
	ModeShorten: "\nEK KURAL: Gereksiz kelimeleri çıkar, metni kısalt ama anlamı koru.",
	ModeFlow:    "\nEK KURAL: Cümle akışını ve bağlaçları iyileştir, okunabilirliği artır.",
	ModeFormal:  "\nEK KURAL: Günlük dili kurumsal/resmi Türkçeye dönüştür.",
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *Client) buildSystemPrompt(plotSummary string, mode ImprovementMode) string {
	prompt := c.systemPrompt
	if mode != ModeFix && mode != "" {
		if extra, ok := modePrompts[mode]; ok {
			prompt += extra
		}
	}
	if plotSummary != "" {
		prompt += "\n\nBAĞLAM (TUTARLILIK REHBERİ — düzeltme yaparken bu bilgiyi referans al):\n" + plotSummary
	}
	return prompt
}

func (c *Client) Improve(text string) (string, error) {
	return c.ImproveWithOptions(text, "", ModeFix)
}

func (c *Client) ImproveWithOptions(text, plotSummary string, mode ImprovementMode) (string, error) {
	c.resetCancel()
	c.streamMu.Lock()
	c.streamText = ""
	c.streaming = true
	c.streamMu.Unlock()
	defer func() {
		c.streamMu.Lock()
		c.streaming = false
		c.streamMu.Unlock()
	}()

	userMessage := "<DÜZELT>\n" + text + "\n</DÜZELT>"
	sysPrompt := c.buildSystemPrompt(plotSummary, mode)

	reqBody := chatRequest{
		Model: "local",
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.15,
		TopP:        0.9,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("istek oluşturulamadı: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		c.endpoint+"/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llama-server'a bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llama-server hatası (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if c.isCancelled() {
			return "", fmt.Errorf("işlem iptal edildi")
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			fullContent.WriteString(delta)
			c.streamMu.Lock()
			c.streamText = fullContent.String()
			c.streamMu.Unlock()
		}
	}

	result := cleanLLMOutput(fullContent.String())
	c.streamMu.Lock()
	c.streamText = result
	c.streamMu.Unlock()
	return result, nil
}

func cleanLLMOutput(raw string) string {
	s := strings.TrimSpace(raw)

	if idx := strings.Index(s, "<DÜZELT>"); idx != -1 {
		s = s[idx+len("<DÜZELT>"):]
	}
	if idx := strings.Index(s, "</DÜZELT>"); idx != -1 {
		s = s[:idx]
	}

	s = strings.TrimSpace(s)

	for _, prefix := range []string{"Düzeltilmiş metin:", "Düzeltilmiş:", "İşte düzeltilmiş metin:", "İşte:"} {
		lower := strings.ToLower(s)
		prefixLower := strings.ToLower(prefix)
		if strings.HasPrefix(lower, prefixLower) {
			s = strings.TrimSpace(s[len(prefix):])
		}
	}

	if len(s) > 0 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	return strings.TrimSpace(s)
}

func (c *Client) GenerateSummary(text string) (string, error) {
	sysPrompt := "Sen bir analiz ve metin özetleme motorusun. Yalnızca istenen özeti oluşturmalısın. Sohbet BAŞLATMA."
	userMsg := "Lütfen aşağıdaki metni dikkatlice oku ve yazarın daha sonra referans alması için 1-2 paragraflık bir olay örgüsü (plot summary) çıkart. Karakterleri ve mekanı tanıt.\nSADECE ÖZETİ YAZ.\n\nMETİN:\n" + text

	reqBody := chatRequest{
		Model: "local",
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.3,
		TopP:        0.9,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("istek oluşturulamadı: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.endpoint+"/v1/chat/completions",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("llama-server'a bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("yanıt okunamadı: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llama-server hatası (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("yanıt ayrıştırılamadı: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llama-server boş yanıt döndü")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

func (c *Client) CheckConsistency(text, plotSummary string) (string, error) {
	if plotSummary == "" {
		return "", fmt.Errorf("tutarlılık denetimi için hikaye özeti gerekli")
	}

	sysPrompt := "Sen bir hikaye tutarlılık denetçisisin. Metin ile bağlam arasındaki çelişkileri bul. Sohbet YAPMA."
	userMsg := fmt.Sprintf("BAĞLAM:\n%s\n\nMETİN:\n%s\n\nÇelişki varsa kısa açıkla. Yoksa 'Tutarlı' yaz.", plotSummary, text)

	reqBody := chatRequest{
		Model: "local",
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.2,
		TopP:        0.9,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(
		c.endpoint+"/v1/chat/completions",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("llama-server'a bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llama-server hatası (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llama-server boş yanıt döndü")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

func (c *Client) IsHealthy() bool {
	resp, err := c.httpClient.Get(c.endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
