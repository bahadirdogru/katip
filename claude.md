# Katip - AI Bağlam Dosyası

Bu dosya, AI asistanların projeyi hızlıca anlaması için optimize edilmiş bağlam sağlar.

## Proje Özeti

Katip, CPU üzerinde yerel LLM ile çalışan bir cross-platform masaüstü Türkçe metin düzenleyicidir. Windows, macOS (Intel & Apple Silicon) ve Linux destekler. Word tarzı "Track Changes" fonksiyonelliği (kelime bazlı diff, editör içi inline markup, onayla/reddet) ile Notion tarzı minimal tasarım sunar. İlk açılışta Setup Wizard eksik bileşenleri algılar ve adım adım yönlendirir. hunspell-wasm ile çevrimdışı Türkçe yazım denetimi yapar. llama-server ve GGUF modeller uygulama içinden otomatik indirilebilir (platforma göre doğru binary seçilir). Header'da kırmızı/sarı/yeşil durum ışığı ile AI sunucu durumu anlık gösterilir. Gece/gündüz modu desteklenir. Yetersiz RAM gibi sorunlar akıllı hata tespitiyle kullanıcıya bildirilir.

## Teknoloji Yığını

- **Backend**: Go 1.25+ / Wails v3 (alpha.74)
- **Frontend**: Svelte 5 (runes: `$state`, `$derived`, `$props`, `$effect`) + TypeScript
- **Editör**: TipTap 2.11 (ProseMirror çekirdeği)
- **Stil**: Tailwind CSS 4 (`@tailwindcss/vite` plugin), `@custom-variant dark` ile class tabanlı dark mode
- **Tasarım**: Word fonksiyonelliği + Notion minimalizmi (hover-to-reveal, ince accent, pastel track changes), gece/gündüz modu
- **LLM**: llama-server subprocess, OpenAI-uyumlu HTTP API (`/v1/chat/completions`)
- **Diff**: `sergi/go-diff` (Go, kelime bazlı), ProseMirror Decoration (frontend inline markup)
- **Yazım Denetimi**: hunspell-wasm (WebAssembly) + tdd-ai/hunspell-tr Türkçe sözlükleri
- **Config**: JSON dosyası, `os.UserConfigDir()/Katip/config.json`
- **Model deposu**: `os.UserConfigDir()/Katip/models/` (GGUF dosyaları)
- **llama-server deposu**: `os.UserConfigDir()/Katip/llama-server/`

## Dizin Yapısı

```
katip/
├── main.go                              # Wails app entry point, embed frontend/dist
├── go.mod                               # module katip, Go 1.25
├── internal/
│   ├── service/katip.go                 # KatipService: Wails'e expose edilen ana API (17 method)
│   ├── llm/
│   │   ├── client.go                    # llama-server HTTP client, <DÜZELT> etiketli prompt, cleanLLMOutput
│   │   ├── manager.go                   # llama-server subprocess yaşam döngüsü + analyzeServerLog (RAM tespiti)
│   │   ├── downloader.go               # GitHub Releases API: llama-server otomatik indirme + zip açma (cross-platform)
│   │   ├── models.go                   # GGUF model kataloğu (IsDefault, MinRAM) + HuggingFace indirme (resume) + GetDefaultModel(), FindModelPartFile()
│   │   ├── signal_unix.go              # Unix (macOS/Linux): SIGTERM ile graceful shutdown (//go:build !windows)
│   │   └── signal_windows.go           # Windows: Process.Kill() ile sonlandırma (//go:build windows)
│   └── diff/engine.go                   # go-diff wrapper: ComputeDiff, ComputeWordDiff (varsayılan)
├── frontend/
│   ├── src/
│   │   ├── App.svelte                   # Ana layout + Setup Wizard + LLM durum ışığı polling + dark mode toggle
│   │   ├── main.ts                      # Svelte 5 mount() entry point
│   │   ├── app.css                      # Tailwind tema (light/dark @custom-variant) + track changes + spell check
│   │   └── lib/
│   │       ├── components/
│   │       │   ├── Editor.svelte        # TipTap editör, hover AI butonu, spell check entegrasyonu
│   │       │   ├── Toolbar.svelte       # Biçimlendirme + belirgin mavi AI İyileştir butonu + applyDiffDecorations
│   │       │   ├── ReviewPanel.svelte   # Sağ panel: Notion tarzı header + separator
│   │       │   ├── ReviewCard.svelte    # Notion tarzı: hover-to-reveal, accent çizgi, kırpılmış diff
│   │       │   ├── SetupWizard.svelte   # İlk açılış kurulum sihirbazı (hoşgeldin, llama, model, tamamlandı)
│   │       │   ├── SpellSuggestion.svelte # Yazım önerisi popup'ı (sağ tık menüsü)
│   │       │   └── SettingsDialog.svelte # Kurulum + model katalogu (varsayılan badge) + hata banner'ları
│   │       ├── editor/
│   │       │   ├── diffDecorations.ts   # ProseMirror Plugin: inline delete + widget insert dekorasyonları
│   │       │   ├── spellChecker.ts      # hunspell-wasm wrapper: init, testSpelling, getSuggestions, tokenize
│   │       │   ├── spellcheckPlugin.ts  # ProseMirror Plugin: wavy underline dekorasyonları (debounced)
│   │       │   └── extensions.ts        # TipTap extension re-exports
│   │       └── stores/
│   │           └── reviewStore.svelte.ts # Svelte 5 $state class: Review CRUD (DİKKAT: .svelte.ts uzantısı zorunlu)
│   ├── public/dictionaries/             # Türkçe hunspell sözlük dosyaları
│   │   ├── tr_TR.aff                    # Affix kuralları (~2.2 MB)
│   │   └── tr_TR.dic                    # Sözlük (~34.5 MB)
│   ├── bindings/                        # Wails otomatik oluşturur (gitignore'da)
│   └── package.json
└── project.md                           # Detaylı mimari doküman
```

## Wails Binding Sistemi

Go'daki public method'lar otomatik olarak frontend'e JS binding olarak sunulur:
- `wails3 generate bindings` komutu `frontend/bindings/` altına JS dosyaları üretir
- Frontend'den import: `import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js'`
- Binding'ler gitignore'dadır, her Go API değişikliğinde yeniden üretilmelidir

### KatipService API (17 Method)

| Method | Parametre | Dönüş | Açıklama |
|--------|-----------|-------|----------|
| `Greet(name)` | string | string | Test metodu |
| `ImproveParagraph(id, text)` | string, string | DiffResult | Paragrafı AI ile iyileştir (kelime bazlı diff) |
| `GetLLMStatus()` | - | map | running, healthy, endpoint, modelPath, lastError |
| `GetServerLog()` | - | string | llama-server stdout/stderr logu |
| `GetConfig()` | - | AppConfig | Uygulama ayarları |
| `UpdateConfig(cfg)` | AppConfig | error | Ayarları güncelle ve kaydet |
| `StartLLMServer()` | - | error | llama-server subprocess başlat |
| `StopLLMServer()` | - | error | llama-server'ı durdur |
| `CheckSetupStatus()` | - | map | İlk açılış durumu: status, llamaInstalled, modelInstalled, vb. |
| `CheckLlamaServer()` | - | map | installed, path, zipExists |
| `DownloadLlamaServer()` | - | error | GitHub'dan llama-server indir (arka plan goroutine) |
| `ReextractLlamaServer()` | - | error | Mevcut zip'ten DLL'ler dahil yeniden çıkar |
| `GetDownloadProgress()` | - | DownloadProgress | llama-server indirme durumu |
| `GetModelCatalog()` | - | []ModelInfo | Mevcut GGUF model listesi (MinRAM, IsDefault dahil) |
| `GetInstalledModels()` | - | []string | İndirilmiş model ID'leri |
| `DownloadModel(modelID)` | string | error | HuggingFace'den model indir (arka plan goroutine) |
| `GetModelDownloadProgress()` | - | DownloadProgress | Model indirme durumu |

### Veri Modelleri

```go
type DiffResult struct {
    ParagraphID string     `json:"paragraphId"`
    Summary     string     `json:"summary"`
    Original    string     `json:"original"`
    Improved    string     `json:"improved"`
    Diffs       []DiffItem `json:"diffs"`
}

type DiffItem struct {
    Type string `json:"type"` // "equal" | "insert" | "delete"
    Text string `json:"text"`
}

type AppConfig struct {
    ModelPath    string `json:"modelPath"`
    ServerBinary string `json:"serverBinary"`
    ServerHost   string `json:"serverHost"`
    ServerPort   int    `json:"serverPort"`
    CtxSize      int    `json:"ctxSize"`
    Threads      int    `json:"threads"`
    SystemPrompt string `json:"systemPrompt"`
}

type DownloadProgress struct {
    Status     string `json:"status"`
    Percent    int    `json:"percent"`     // 0-100, -1 hata durumunda
    Downloaded int64  `json:"downloaded"`
    Total      int64  `json:"total"`
    Error      string `json:"error,omitempty"`
}

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
```

### Yerleşik Model Kataloğu

| ID | Model | Boyut | Min RAM | Dil | Varsayılan |
|----|-------|-------|---------|-----|------------|
| `turkcell-7b-q4km` | Turkcell-LLM-7b-v1 Q4_K_M | ~4.5 GB | 8 GB | Türkçe | Evet |
| `openr1-qwen-7b-tr-q4km` | OpenR1-Qwen-7B-Turkish Q4_K_M | ~4.5 GB | 8 GB | Türkçe | - |
| `qwen25-3b-q4km` | Qwen2.5-3B-Instruct Q4_K_M | ~2.0 GB | 4 GB | Çok dilli | - |
| `bitnet-2b-4t` | BitNet b1.58-2B-4T | ~1.1 GB | 2 GB | İngilizce | - |

## Veri Akışı

```
== AI İyileştirme Akışı ==
Kullanıcı -> [AI İyileştir butonuna tıklar (Toolbar veya hover)]
  -> Editor/Toolbar: paragraf textContent alır
  -> Wails binding: ImproveParagraph(id, text)
  -> Go KatipService: llmClient.Improve(text)
  -> HTTP POST -> llama-server /v1/chat/completions
     (metin <DÜZELT> etiketi içinde, temperature=0.15)
  -> Go: cleanLLMOutput(yanıt) -- etiket/önek temizle
  -> Go: diffEngine.ComputeWordDiff(original, improved) -- kelime bazlı
  -> DiffResult JSON -> Frontend
  -> reviewStore.addReview(result)
  -> applyDiffDecorations() -> editör içi inline markup göster
     (delete: kırmızı üstü çizili, insert: yeşil altı çizili widget)
  -> ReviewPanel: Notion tarzı kart göster (hover-to-reveal butonlar)
  -> Kullanıcı Onayla: clearDecorations() + tr.replaceWith()
  -> Kullanıcı Reddet: clearDecorations() + reviewStore.rejectReview()

== llama-server Hata Tespiti ==
llama-server subprocess çöktüğünde:
  -> manager.go: logBuf'tan analyzeServerLog() çağrılır
  -> Bilinen hata kalıpları:
     "failed to allocate" -> BELLEK_YETERSIZ (RAM uyarısı)
     "not a valid gguf"   -> MODEL_BOZUK
     "address already in use" -> PORT_KULLANILIYOR
  -> Frontend: errorType'a göre renkli banner gösterilir
     (RAM hatası: kırmızı kutu + küçük model önerisi)

== llama-server İndirme Akışı (Cross-Platform) ==
SettingsDialog: "İndir" butonuna tıkla
  -> Wails binding: DownloadLlamaServer()
  -> Go goroutine: GitHub Releases API -> en son sürümü bul
  -> findAssetName(): runtime.GOOS + runtime.GOARCH ile platform tespiti
     Windows: win-cpu-x64 / win-cpu-arm64 -> .exe + .dll
     macOS:   mac-arm64 / mac-x64         -> binary + .dylib + .metal
     Linux:   ubuntu-x64 / ubuntu-arm64   -> binary + .so
  -> HTTP GET -> zip indir -> zip aç -> platforma uygun dosyalar çıkar
  -> Kayıt: os.UserConfigDir()/Katip/llama-server/
  -> config.ServerBinary otomatik güncellenir
  -> Frontend: setInterval ile GetDownloadProgress() poll

== GGUF Model İndirme Akışı ==
SettingsDialog: model kartındaki "İndir" butonuna tıkla
  -> Wails binding: DownloadModel(modelID)
  -> Go goroutine: HuggingFace resolve URL'den HTTP GET
  -> .part dosyasına yaz (resume desteği: Range header)
  -> Tamamlanınca .part -> .gguf yeniden adlandır
  -> Kayıt: os.UserConfigDir()/Katip/models/<filename>.gguf
  -> config.ModelPath otomatik güncellenir
  -> Frontend: setInterval ile GetModelDownloadProgress() poll

== İlk Açılış Setup Wizard Akışı ==
App.svelte: onMount
  -> Wails binding: CheckSetupStatus()
  -> Go: llama-server var mı? zip var mı? varsayılan model var mı? .part var mı?
  -> Go: Config boşsa ama dosyalar mevcutsa -> otomatik config doldur + saveConfig()
  -> Dönüş: status ("ready" | "zip_found" | "llama_missing" | "model_partial" | "model_missing")
  -> status === "ready" -> wizard atlanır, ana editör gösterilir
  -> status !== "ready" -> SetupWizard.svelte gösterilir
     Adım 1: Hoşgeldiniz (eksik bileşenlerin özeti)
     Adım 2: llama-server kurulumu (zip_found -> çıkar, llama_missing -> indir)
     Adım 3: Model indirme (model_partial -> devam et, model_missing -> indir)
     Adım 4: Tamamlandı -> "Kullanmaya Başla" -> CheckSetupStatus() tekrar çağrılır

== Header Durum Işığı ==
App.svelte: onMount + setInterval(5000ms)
  -> Wails binding: GetLLMStatus()
  -> Dönüş: { running, healthy, ... }
  -> healthy=true  -> yeşil ışık ("AI Hazır")
  -> running=true  -> sarı ışık, animasyonlu ping ("Yükleniyor")
  -> else          -> kırmızı ışık ("AI Kapalı")
  -> Işığa tıklanınca SettingsDialog açılır

== Gece/Gündüz Modu ==
App.svelte header'da ay/güneş ikonu
  -> toggleTheme(): document.documentElement.classList.toggle('dark')
  -> localStorage.setItem('katip-theme', 'dark'|'light')
  -> onMount: localStorage'dan oku, uygula
  -> CSS: @custom-variant dark -> .dark sınıfı ile tema değişkenleri override

== Türkçe Yazım Denetimi Akışı ==
Editor.svelte: onMount
  -> spellChecker.ts: initSpellChecker()
     -> fetch("/dictionaries/tr_TR.aff") + fetch("/dictionaries/tr_TR.dic")
     -> hunspell-wasm: createHunspellFromStrings(aff, dic)
  -> spellcheckPlugin.ts: ProseMirror Plugin kayıt
     -> Her doc değişikliğinde debounce (400ms)
     -> buildSpellDecorations(): tüm text block'ları tara
     -> tokenize() -> testSpelling() -> Decoration.inline("spell-error")
  -> Kullanıcı sağ tık -> handleContextMenu()
     -> Tıklanan pozisyonda spell-error dekorasyonu var mı?
     -> SpellSuggestion.svelte popup göster
     -> getSuggestions(word) -> öneri listesi
     -> Seçim: editor'da kelimeyi değiştir + spell dekorasyonlarını yenile
     -> "Sözlüğe ekle": hunspell.addWord() + popup kapat
```

## LLM İletişim Detayları

**Sistem Prompt**: Metin düzeltme motoru olarak tanımlanır. Sohbet/soru/açıklama yasaklanır.
- Kullanıcı mesajı `<DÜZELT>metin</DÜZELT>` etiketi içinde gönderilir
- `temperature: 0.15`, `top_p: 0.9` -- düşük yaratıcılık, tutarlı düzeltme
- `cleanLLMOutput()`: Etiketler, önek'ler ("Düzeltilmiş metin:"), tırnak işaretleri otomatik temizlenir

## Svelte 5 Kuralları

Bu projede Svelte 5 runes kullanılır. Svelte 4 syntax'ı kullanma:
- Durum: `let x = $state(value)` (export let yerine `$props()`)
- Türetilmiş: `let y = $derived(expr)`
- Efekt: `$effect(() => { ... })`
- Props: `let { prop1, prop2 }: Props = $props()`
- Event handler: `onclick={fn}` (on:click yerine)
- Mount: `mount(App, { target })` (new App() yerine)
- Store: class içinde `$state` kullan, Svelte store API'si yerine
- **`$` önek yasağı**: Svelte 5 rune modunda `$` ile başlayan değişken adı kullanılamaz (ProseMirror `$from` gibi alanlar `const resolved = state.selection.$from` şeklinde erişilmeli)

**KRİTİK**: `$state` rune kullanan TypeScript dosyaları `.svelte.ts` uzantısına sahip olmalıdır. Düz `.ts` dosyalarında rune'lar Svelte derleyicisi tarafından işlenmez ve runtime hatası verir.

## Tailwind CSS 4 Tema (Light/Dark)

`app.css` içinde `@custom-variant dark (&:where(.dark, .dark *))` ile class tabanlı dark mode.

**Light tema** (`@theme` bloğu):
- `--color-primary` / `--color-primary-dark`: Mavi (#2563eb / #1d4ed8)
- `--color-diff-insert-bg` / `--color-diff-insert-text`: Pastel yeşil (rgba(0,128,0,0.08) / #27ae60)
- `--color-diff-delete-bg` / `--color-diff-delete-text`: Pastel kırmızı (rgba(255,0,0,0.08) / #c0392b)
- `--color-surface` / `--color-surface-secondary`: Beyaz / Açık gri (#f9fafb)
- `--color-border`: Gri kenarlık (#e5e7eb)
- `--color-text-primary` / `--color-text-secondary`: Koyu / Açık metin
- `--color-accent-blue` / `--color-accent-green` / `--color-accent-gray`: Review kartı accent çizgi renkleri

**Dark tema** (`.dark body` CSS override ile aynı değişkenleri geçersiz kılar):
- `--color-surface`: #1a1a2e, `--color-surface-secondary`: #16213e
- `--color-border`: #2d3748, `--color-text-primary`: #e2e8f0, `--color-text-secondary`: #718096

**Dark mode geçişi**: `<html>` elemanına `.dark` class eklenir/çıkarılır, tercih `localStorage('katip-theme')` ile saklanır.

Özel CSS sınıfları: `.review-card`, `.review-card-accent-*`, `.review-card-actions` (hover-to-reveal), `.review-action-btn`, `.review-separator`

## Gereksinimler (Platform Bazlı)

| Platform | Gereksinimler |
|----------|---------------|
| Tümü | Go 1.25+, Node.js 18+, Wails v3 CLI |
| Windows | Windows 10+ (WebView2, Win11'de dahili) |
| macOS | macOS 10.15+, Xcode Command Line Tools (`xcode-select --install`) |
| Linux (Debian/Ubuntu) | `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `gcc` |
| Linux (Fedora) | `gtk3-devel`, `webkit2gtk4.1-devel`, `gcc` |
| Linux (Arch) | `gtk3`, `webkit2gtk-4.1`, `gcc` |

## Geliştirme Komutları

```bash
cd katip

# Geliştirme modu (hot reload) -- tüm platformlar
wails3 dev

# Binding'leri yeniden oluştur (Go API değişince)
wails3 generate bindings

# Frontend build
cd frontend && npm run build
```

### Windows

```powershell
# Zamanlama sorunu varsa: önce vite başlat, sonra Go binary çalıştır
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -ldflags "-H windowsgui" -o bin/katip.exe .
$env:FRONTEND_DEVSERVER_URL="http://localhost:9245"; .\bin\katip.exe

# Go build (syso yoksa -ldflags gerekli)
go build -ldflags "-H windowsgui" -o bin/katip.exe .
```

### macOS

```bash
# Native build
go build -o bin/katip .

# Universal binary (Intel + Apple Silicon)
wails3 task darwin:build:universal

# .app bundle oluştur
wails3 task darwin:package

# Zamanlama sorunu varsa
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -o bin/katip .
FRONTEND_DEVSERVER_URL="http://localhost:9245" ./bin/katip
```

### Linux

```bash
# Native build (CGO gerekli)
CGO_ENABLED=1 go build -o bin/katip .

# AppImage oluştur
wails3 task linux:create:appimage

# deb/rpm paketleri
wails3 task linux:create:deb
wails3 task linux:create:rpm

# Zamanlama sorunu varsa
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -o bin/katip .
FRONTEND_DEVSERVER_URL="http://localhost:9245" ./bin/katip
```

### Cross-Compile (Docker)

```bash
# Docker imajını kur (ilk seferde)
wails3 task setup:docker

# Herhangi bir platformdan hedef platforma derle
wails3 task darwin:build ARCH=arm64
wails3 task linux:build ARCH=amd64

# Tam paket (mevcut platform)
wails3 build
```

## Dosya Depolama Konumları

Tüm veriler `os.UserConfigDir()/Katip/` altında saklanır:

| Veri | Alt Dizin |
|------|-----------|
| Uygulama ayarları | `config.json` |
| llama-server binary + kütüphaneler | `llama-server/` (Win: .exe+.dll, macOS: binary+.dylib+.metal, Linux: binary+.so) |
| llama-server indirme arşivi | `llama-server/*.zip` (tekrar açmak için saklanır) |
| GGUF model dosyaları | `models/<model>.gguf` |
| İndirme geçici dosyaları | `models/<model>.gguf.part` |

**Platform bazlı kök dizin:**

| Platform | `os.UserConfigDir()` Sonucu |
|----------|----------------------------|
| Windows | `C:\Users\<user>\AppData\Roaming\Katip\` |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

## Cross-Platform Notlar

- **Subprocess sonlandırma**: macOS/Linux'ta SIGTERM (graceful), Windows'ta Process.Kill() -- `signal_unix.go` / `signal_windows.go`
- **WebView**: Windows → WebView2 (Edge), macOS → WebKit (dahili), Linux → WebKitGTK 4.1
- **llama-server asset seçimi**: `downloader.go` → `findAssetName()` fonksiyonu `runtime.GOOS` + `runtime.GOARCH` ile platforma uygun GitHub release asset'ini seçer
- **Arşiv çıkarma**: `.exe`, `.dll`, `.so`, `.dylib`, `.metal`, `LICENSE` dosyaları tutulur; binary'ye `GetLlamaServerPath()` ile erişilir (Windows'ta `.exe` uzantısı otomatik eklenir)
- **macOS build**: CGO gerekli, `MACOSX_DEPLOYMENT_TARGET=10.15`. Universal binary (Intel+ARM) `lipo` ile oluşturulur
- **Linux build**: CGO gerekli, `libgtk-3-dev` + `libwebkit2gtk-4.1-dev` bağımlılıkları
- **Windows build**: `CGO_ENABLED=0`, syso yoksa `-ldflags "-H windowsgui"` gerekli
- **Docker cross-compile**: `build/docker/Dockerfile.cross` ile herhangi bir platformdan darwin/linux/windows hedeflerine derleme

## Bilinen Sınırlamalar ve Gelecek Çalışmalar

- Dosya açma/kaydetme henüz yok
- Streaming token desteği henüz yok (şu an tam yanıt bekleniyor)
- Birden fazla paragraf eşzamanlı iyileştirme henüz desteklenmiyor
- Sözlüğe eklenen kelimeler oturum bazlıdır (kalıcı değil)
- `wails3 dev` bazen zamanlama sorunu yaşayabilir (vite dev server Go binary'den sonra hazır olamaz)
- Windows'ta syso dosyası yoksa `go build` için `-ldflags "-H windowsgui"` gereklidir
- Vite build sırasında hunspell-wasm `fs/promises` ve `module` uyarıları gösterir (işlevselliği etkilemez)
