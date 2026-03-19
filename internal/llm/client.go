package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	endpoint     string
	httpClient   *http.Client
	systemPrompt string
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

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
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

func (c *Client) Improve(text string) (string, error) {
	userMessage := "<DÜZELT>\n" + text + "\n</DÜZELT>"

	reqBody := chatRequest{
		Model: "local",
		Messages: []chatMessage{
			{Role: "system", Content: c.systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.15,
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

	result := cleanLLMOutput(chatResp.Choices[0].Message.Content)
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

func (c *Client) IsHealthy() bool {
	resp, err := c.httpClient.Get(c.endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
