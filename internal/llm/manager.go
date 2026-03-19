package llm

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	cmd       *exec.Cmd
	mu        sync.Mutex
	running   bool
	lastError string
	logBuf    *ringBuffer
}

func NewManager() *Manager {
	return &Manager{
		logBuf: newRingBuffer(8192),
	}
}

type ServerConfig struct {
	BinaryPath string
	ModelPath  string
	Host       string
	Port       int
	CtxSize    int
	Threads    int
}

func DefaultConfig() ServerConfig {
	return ServerConfig{
		Host:    "127.0.0.1",
		Port:    8089,
		CtxSize: 4096,
		Threads: 4,
	}
}

func (m *Manager) Start(cfg ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("llama-server zaten çalışıyor")
	}

	if _, err := os.Stat(cfg.BinaryPath); err != nil {
		return fmt.Errorf("llama-server binary bulunamadı: %s", cfg.BinaryPath)
	}
	if _, err := os.Stat(cfg.ModelPath); err != nil {
		return fmt.Errorf("model dosyası bulunamadı: %s", cfg.ModelPath)
	}

	m.lastError = ""
	m.logBuf.Reset()

	args := []string{
		"-m", cfg.ModelPath,
		"--host", cfg.Host,
		"--port", fmt.Sprintf("%d", cfg.Port),
		"-c", fmt.Sprintf("%d", cfg.CtxSize),
		"-t", fmt.Sprintf("%d", cfg.Threads),
	}

	m.cmd = exec.Command(cfg.BinaryPath, args...)
	m.cmd.Stdout = m.logBuf
	m.cmd.Stderr = m.logBuf

	if err := m.cmd.Start(); err != nil {
		m.lastError = err.Error()
		return fmt.Errorf("llama-server başlatılamadı: %w", err)
	}

	m.running = true
	slog.Info("llama-server başlatıldı", "pid", m.cmd.Process.Pid, "port", cfg.Port, "model", cfg.ModelPath)

	go func() {
		err := m.cmd.Wait()
		m.mu.Lock()
		m.running = false
		if err != nil {
			friendlyErr := m.analyzeServerLog(err)
			m.lastError = friendlyErr
			slog.Error("llama-server kapandı", "error", friendlyErr)
		} else {
			slog.Info("llama-server kapandı")
		}
		m.mu.Unlock()
	}()

	go m.waitForStartup(cfg, 60*time.Second)

	return nil
}

func (m *Manager) waitForStartup(cfg ServerConfig, timeout time.Duration) {
	endpoint := fmt.Sprintf("http://%s:%d/health", cfg.Host, cfg.Port)
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		m.mu.Lock()
		if !m.running {
			m.mu.Unlock()
			return
		}
		m.mu.Unlock()

		resp, err := client.Get(endpoint)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				slog.Info("llama-server hazır", "port", cfg.Port)
				return
			}
		}
		time.Sleep(2 * time.Second)
	}

	m.mu.Lock()
	if m.running {
		m.lastError = fmt.Sprintf("Sunucu %v içinde hazır olmadı", timeout)
		slog.Warn("llama-server startup timeout", "timeout", timeout)
	}
	m.mu.Unlock()
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running || m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	slog.Info("llama-server durduruluyor...")

	if err := terminateProcess(m.cmd.Process); err != nil {
		if killErr := m.cmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("süreç sonlandırılamadı: %w", killErr)
		}
	}

	done := make(chan struct{})
	go func() {
		m.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		slog.Warn("llama-server zorla kapatıldı")
	}

	m.running = false
	return nil
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *Manager) LastError() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastError
}

func (m *Manager) Log() string {
	return m.logBuf.String()
}

type ringBuffer struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	max  int
}

func newRingBuffer(max int) *ringBuffer {
	return &ringBuffer{max: max}
}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, err = r.buf.Write(p)
	if r.buf.Len() > r.max {
		data := r.buf.Bytes()
		trimmed := data[len(data)-r.max:]
		r.buf.Reset()
		r.buf.Write(trimmed)
	}
	return n, err
}

func (r *ringBuffer) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func (r *ringBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf.Reset()
}

func (m *Manager) analyzeServerLog(processErr error) string {
	log := strings.ToLower(m.logBuf.String())

	if strings.Contains(log, "failed to allocate") ||
		strings.Contains(log, "unable to allocate") ||
		strings.Contains(log, "out of memory") ||
		strings.Contains(log, "alloc_tensor_range: failed") {
		return "BELLEK_YETERSIZ: Model için yeterli RAM yok. Daha küçük bir model seçin (ör. Qwen2.5-3B ~2 GB veya BitNet-2B ~1 GB)."
	}

	if strings.Contains(log, "failed to load model") ||
		strings.Contains(log, "failed to load model") {
		if strings.Contains(log, "not a valid gguf file") ||
			strings.Contains(log, "invalid magic") {
			return "MODEL_BOZUK: Model dosyası bozuk veya geçersiz. Modeli silip yeniden indirin."
		}
		return "MODEL_YÜKLENEMEDI: Model dosyası yüklenemedi. Dosyanın doğru bir GGUF model olduğundan emin olun."
	}

	if strings.Contains(log, "address already in use") ||
		strings.Contains(log, "bind failed") {
		return "PORT_KULLANILIYOR: Sunucu portu zaten başka bir uygulama tarafından kullanılıyor. Portu değiştirin veya diğer uygulamayı kapatın."
	}

	if strings.Contains(log, "model file not found") ||
		strings.Contains(log, "no such file") {
		return "MODEL_BULUNAMADI: Model dosyası belirtilen yolda bulunamadı. Dosya yolunu kontrol edin."
	}

	return fmt.Sprintf("Süreç sonlandı: %v", processErr)
}
