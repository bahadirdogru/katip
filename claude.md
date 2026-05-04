# Katip - AI Baglam Dosyasi

Bu dosya, AI asistanlarin projeyi hizlica anlamasi icin optimize edilmis baglam saglar. Kusursuz entegrasyon amaciyla tum Cross Platform derleme ve UI renk seceneklerini de icerir. Emojilerden arindirilmistir.
REFERANSLAR: [project.md: Detayli mimari kararlar, README.md: Pazarlama vitrini]

## YENI MIMARI VE VIZYON (EPICS)
Yerel LLM destekli masaustu Turkce metin duzenleyici. Go 1.25+/Wails v3/Svelte 5/Tailwind 4. Tamamen cevrimdisi (gizlilik odakli).
- **.kitap Formatı**: Metin (MD), meta veriler, eklenecek medya/fontlar, RAG (olay ozeti) paneli ve diff gecmisini barindiran asenkron (e-posta) ZIP/JSON paket formatidir.
- **Git-Diff Gecmisi**: Klasik Word Track Changes yerine satır bazli "Isim-Tarih" log (Merge conflict harictir).
- **RAG Tutarlilik Analizi**: AI, zaman hatalarini (.orn: karakterin goz rengi degisimi) bulmak icin .kitap icindeki Olay/Karakter Ozeti tablosunu statik referans alir.
- **.tarz Profil Dosyalari**: Yayinevi jargonu ve standart kelime kurallari import/export edilebilir. Aktif bolum taranip Oneri Kartlarina donusturulur.

## Proje Ozeti
Katip, CPU uzerinde yerel LLM ile calisan bir cross-platform masaustu Turkce metin duzenleyicidir. Windows, macOS ve Linux destekler. Word tarzi "Track Changes" fonksiyonelligi ile Notion tarzi minimal tasarim sunar. Ilk acilista Setup Wizard eksik bilesenleri algilar ve yonlendirir. hunspell-wasm ile cevrimdisi yazim denetimi yapar. llama-server ve GGUF modeller uygulama icinden otomatik indirilebilir. Header'da kirmizi/sari/yesil isik ile AI sunucu durumu gosterilir. Gece/gunduz modu desteklenir. Yetersiz RAM uyarisi verilir.

## Teknoloji Yigini
- **Backend**: Go 1.25+ / Wails v3 (alpha.74)
- **Frontend**: Svelte 5 (runes: `$state`, `$derived`, `$props`, `$effect`) + TypeScript
- **Editor**: TipTap 2.11 (ProseMirror cekirdegi)
- **Stil**: Tailwind CSS 4 (`@tailwindcss/vite` plugin), `@custom-variant dark` ile class tabanli dark mode
- **Tasarim**: Word fonksiyonelligi + Notion minimalizmi (hover-to-reveal, ince accent, pastel track changes), gece/gunduz modu
- **LLM**: llama-server subprocess, OpenAI-uyumlu HTTP API (`/v1/chat/completions`)
- **Diff**: `sergi/go-diff` (Go, kelime bazli), ProseMirror Decoration
- **Yazim Denetimi**: hunspell-wasm (WebAssembly) + tdd-ai/hunspell-tr Turkce sozlukleri
- **Config**: JSON dosyasi, `os.UserConfigDir()/Katip/config.json`
- **Model deposu**: `os.UserConfigDir()/Katip/models/` (GGUF dosyalari)
- **llama-server deposu**: `os.UserConfigDir()/Katip/llama-server/`

## Dizin Yapisi
```
katip/
├── main.go                              # Wails app entry point, embed frontend/dist
├── go.mod                               # module katip, Go 1.25
├── internal/
│   ├── service/katip.go                 # KatipService: Wails'e expose edilen ana API (17 method)
│   ├── llm/
│   │   ├── client.go                    # llama-server HTTP client, <DUZELT> etiketli prompt, cleanLLMOutput
│   │   ├── manager.go                   # llama-server subprocess yasam dongusu + analyzeServerLog (RAM tespiti)
│   │   ├── downloader.go               # GitHub Releases API: llama-server otomatik indirme + zip acma (cross-platform)
│   │   ├── models.go                   # GGUF model katalogu + HuggingFace indirme (resume)
│   │   ├── signal_unix.go              # Unix (macOS/Linux): SIGTERM ile graceful shutdown
│   │   └── signal_windows.go           # Windows: Process.Kill() ile sonlandirma
│   └── diff/engine.go                   # go-diff wrapper: ComputeDiff, ComputeWordDiff
├── frontend/
│   ├── src/
│   │   ├── App.svelte                   # Ana layout + Setup Wizard + LLM polling + dark mode toggle
│   │   ├── main.ts                      # Svelte 5 mount() entry point
│   │   ├── app.css                      # Tailwind tema (light/dark @custom-variant) + track changes + spell check
│   │   └── lib/
│   │       ├── components/
│   │       │   ├── Editor.svelte        # TipTap editor, hover AI butonu, spell check entegrasyonu
│   │       │   ├── Toolbar.svelte       # Bicimlendirme + belirgin mavi AI Iyilestir butonu + applyDiffDecorations
│   │       │   ├── ReviewPanel.svelte   # Sag panel: Notion tarzi header + separator
│   │       │   ├── ReviewCard.svelte    # hover-to-reveal, accent cizgi, kirpilmis diff
│   │       │   ├── SetupWizard.svelte   # Ilk acilis kurulum sihirbazi
│   │       │   ├── SpellSuggestion.svelte # Yazim onerisi popup'i (sag tik)
│   │       │   └── SettingsDialog.svelte # Kurulum + model katalogu + hata banner'lari
│   │       ├── editor/
│   │       │   ├── diffDecorations.ts   # ProseMirror Plugin: inline delete + widget insert
│   │       │   ├── spellChecker.ts      # hunspell-wasm wrapper: init, testSpelling, getSuggestions
│   │       │   ├── spellcheckPlugin.ts  # ProseMirror Plugin: wavy underline (debounced)
│   │       │   └── extensions.ts        # TipTap extension re-exports
│   │       └── stores/
│   │           └── reviewStore.svelte.ts # Svelte 5 $state class: Review CRUD (DIKKAT: .svelte.ts uzantisi zorunlu)
│   ├── public/dictionaries/             # Turkce hunspell sozluk dosyalari
│   │   ├── tr_TR.aff                    # Affix kurallari
│   │   └── tr_TR.dic                    # Sozluk
│   ├── bindings/                        # Wails otomatik olusturur (gitignore'da)
│   └── package.json
└── project.md                           # Detayli mimari dokuman
```

## Wails Binding Sistemi
Go'daki public method'lar otomatik olarak frontend'e JS binding olarak sunulur:
- `wails3 generate bindings` komutu `frontend/bindings/` altina JS dosyalari uretir
- Frontend'den import: `import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js'`
- Binding'ler gitignore'dadir, her Go API degisikliginde yeniden uretilmelidir

### KatipService API (17 Method)
| Method | Parametre | Donus | Aciklama |
|--------|-----------|-------|----------|
| `Greet(name)` | string | string | Test metodu |
| `ImproveParagraph(id, text)` | string, string | DiffResult | Paragrafi AI ile iyilestir (kelime bazli diff) |
| `GetLLMStatus()` | - | map | running, healthy, endpoint, modelPath, lastError |
| `GetServerLog()` | - | string | llama-server stdout/stderr logu |
| `GetConfig()` | - | AppConfig | Uygulama ayarlari |
| `UpdateConfig(cfg)` | AppConfig | error | Ayarlari guncelle ve kaydet |
| `StartLLMServer()` | - | error | llama-server subprocess baslat |
| `StopLLMServer()` | - | error | llama-server'i durdur |
| `CheckSetupStatus()` | - | map | Ilk acilis durumu: status, llamaInstalled, modelInstalled, vb. |
| `CheckLlamaServer()` | - | map | installed, path, zipExists |
| `DownloadLlamaServer()` | - | error | GitHub'dan llama-server indir (arka plan goroutine) |
| `ReextractLlamaServer()` | - | error | Mevcut zip'ten DLL'ler dahil yeniden cikar |
| `GetDownloadProgress()` | - | DownloadProgress | llama-server indirme durumu |
| `GetModelCatalog()` | - | []ModelInfo | Mevcut GGUF model listesi (MinRAM, IsDefault dahil) |
| `GetInstalledModels()` | - | []string | Indirilmis model ID'leri |
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
    Percent    int    `json:"percent"`
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

### Yerlesik Model Katalogu
| ID | Model | Boyut | Min RAM | Dil | Varsayilan |
|----|-------|-------|---------|-----|------------|
| `turkcell-7b-q4km` | Turkcell-LLM-7b-v1 Q4_K_M | ~4.5 GB | 8 GB | Turkce | Evet |
| `openr1-qwen-7b-tr-q4km` | OpenR1-Qwen-7B-Turkish Q4_K_M | ~4.5 GB | 8 GB | Turkce | - |
| `qwen25-3b-q4km` | Qwen2.5-3B-Instruct Q4_K_M | ~2.0 GB | 4 GB | Cok dilli | - |
| `bitnet-2b-4t` | BitNet b1.58-2B-4T | ~1.1 GB | 2 GB | Ingilizce | - |

## Veri Akisi
```
== AI Iyilestirme Akisi ==
Kullanici -> [AI Iyilestir butonuna tiklar (Toolbar veya hover)]
  -> Editor/Toolbar: paragraf textContent alir
  -> Wails binding: ImproveParagraph(id, text)
  -> Go KatipService: llmClient.Improve(text)
  -> HTTP POST -> llama-server /v1/chat/completions (metin <DUZELT> etiketi icinde, temperature=0.15)
  -> Go: cleanLLMOutput(yanit) -- etiket/onek temizle
  -> Go: diffEngine.ComputeWordDiff(original, improved) -- kelime bazli
  -> DiffResult JSON -> Frontend
  -> reviewStore.addReview(result)
  -> applyDiffDecorations() -> editor ici inline markup goster (delete: kirmizi ustu cizili, insert: yesil alti cizili widget)
  -> ReviewPanel: Notion tarzi kart goster (hover-to-reveal butonlar)
  -> Kullanici Onayla: clearDecorations() + tr.replaceWith()
  -> Kullanici Reddet: clearDecorations() + reviewStore.rejectReview()

== llama-server Hata Tespiti ==
llama-server subprocess coktugunde:
  -> manager.go: logBuf'tan analyzeServerLog() cagirilir
  -> Bilinen hata kaliplari: "failed to allocate" -> BELLEK_YETERSIZ (RAM uyarisi), "not a valid gguf" -> MODEL_BOZUK, "address already in use" -> PORT_KULLANILIYOR
  -> Frontend: errorType'a gore renkli banner gosterilir

== llama-server Indirme Akisi (Cross-Platform) ==
SettingsDialog: "Indir" butonuna tikla
  -> Wails binding: DownloadLlamaServer()
  -> Go goroutine: GitHub Releases API -> en son surumu bul
  -> findAssetName(): runtime.GOOS + runtime.GOARCH ile platform tespiti
     Windows: win-cpu-x64 / win-cpu-arm64 -> .exe + .dll
     macOS:   mac-arm64 / mac-x64         -> binary + .dylib + .metal
     Linux:   ubuntu-x64 / ubuntu-arm64   -> binary + .so
  -> HTTP GET -> zip indir -> zip ac -> platforma uygun dosyalari cikar
  -> Kayit: os.UserConfigDir()/Katip/llama-server/
  -> config.ServerBinary otomatik guncellenir
  -> Frontend: setInterval ile GetDownloadProgress() poll

== GGUF Model Indirme Akisi ==
SettingsDialog: model kartindaki "Indir" butonuna tikla
  -> Wails binding: DownloadModel(modelID)
  -> Go goroutine: HuggingFace resolve URL'den HTTP GET
  -> .part dosyasina yaz (resume destegi: Range header)
  -> Tamamlaninca .part -> .gguf yeniden adlandir
  -> Kayit: os.UserConfigDir()/Katip/models/<filename>.gguf
  -> config.ModelPath otomatik guncellenir
  -> Frontend: setInterval ile GetModelDownloadProgress() poll

== Ilk Acilis Setup Wizard Akisi ==
App.svelte: onMount
  -> Wails binding: CheckSetupStatus()
  -> Go: llama-server var mi? zip var mi? varsayilan model var mi? .part var mi?
  -> Go: Config bossa ama dosyalar mevcutsa -> otomatik config doldur + saveConfig()
  -> Donus: status ("ready" | "zip_found" | "llama_missing" | "model_partial" | "model_missing")
  -> status === "ready" -> wizard atlanir, ana editor gosterilir
  -> status !== "ready" -> SetupWizard.svelte gosterilir
     Adim 1: Hosgeldiniz
     Adim 2: llama-server kurulumu (zip_found -> cikar, llama_missing -> indir)
     Adim 3: Model indirme (model_partial -> devam et, model_missing -> indir)
     Adim 4: Tamamlandi -> "Kullanmaya Basla" -> CheckSetupStatus() tekrar cagrilir

== Header Durum Isigi ==
App.svelte: onMount + setInterval(5000ms)
  -> Wails binding: GetLLMStatus()
  -> Donus: { running, healthy, ... }
  -> healthy=true  -> yesil isik ("AI Hazir")
  -> running=true  -> sari isik, animasyonlu ping ("Yukleniyor")
  -> else          -> kirmizi isik ("AI Kapali")
  -> Isiga tiklaninca SettingsDialog acilir

== Gece/Gunduz Modu ==
App.svelte header'da ay/gunes ikonu
  -> toggleTheme(): document.documentElement.classList.toggle('dark')
  -> localStorage.setItem('katip-theme', 'dark'|'light')
  -> onMount: localStorage'dan oku, uygula
  -> CSS: @custom-variant dark -> .dark sinifi ile tema degiskenleri override

== Turkce Yazim Denetimi Akisi ==
Editor.svelte: onMount
  -> spellChecker.ts: initSpellChecker()
     -> fetch("/dictionaries/tr_TR.aff") + fetch("/dictionaries/tr_TR.dic")
     -> hunspell-wasm: createHunspellFromStrings(aff, dic)
  -> spellcheckPlugin.ts: ProseMirror Plugin kayit
     -> Her doc degisikliginde debounce (400ms)
     -> buildSpellDecorations(): tum text block'lari tara
     -> tokenize() -> testSpelling() -> Decoration.inline("spell-error")
  -> Kullanici sag tik -> handleContextMenu()
     -> Tiklanan pozisyonda spell-error dekorasyonu var mi?
     -> SpellSuggestion.svelte popup goster
     -> getSuggestions(word) -> oneri listesi
     -> Secim: editor'da kelimeyi degistir + spell dekorasyonlarini yenile
     -> "Sozluge ekle": hunspell.addWord() + popup kapat
```

## LLM Iletisim Detaylari
**Sistem Prompt**: Metin duzeltme motoru olarak tanimlanir. Sohbet/soru/aciklama yasaklanir.
- Kullanici mesaji `<DUZELT>metin</DUZELT>` etiketi icinde gonderilir
- `temperature: 0.15`, `top_p: 0.9` -- dusuk yaraticilik, tutarli duzeltme
- `cleanLLMOutput()`: Etiketler, onek'ler ("Duzeltilmis metin:"), tirnak isaretleri otomatik temizlenir

## Svelte 5 Kurallari
Bu projede Svelte 5 runes kullanilir. Svelte 4 syntax'i kullanma:
- Durum: `let x = $state(value)` (export let yerine `$props()`)
- Turetilmis: `let y = $derived(expr)`
- Efekt: `$effect(() => { ... })`
- Props: `let { prop1, prop2 }: Props = $props()`
- Event handler: `onclick={fn}` (on:click yerine)
- Mount: `mount(App, { target })` (new App() yerine)
- Store: class icinde `$state` kullan, Svelte store API'si yerine
- **`$` onek yasagi**: Svelte 5 rune modunda `$` ile baslayan degisken adi kullanilamaz (ProseMirror `$from` gibi alanlar `const resolved = state.selection.$from` seklinde erisilmeli)
**KRITIK**: `$state` rune kullanan TypeScript dosyalari `.svelte.ts` uzantisina sahip olmalidir. Duz `.ts` dosyalarinda rune'lar Svelte derleyicisi tarafindan islenmez ve runtime hatasi verir.

## Tailwind CSS 4 Tema (Light/Dark)
`app.css` icinde `@custom-variant dark (&:where(.dark, .dark *))` ile class tabanli dark mode.
**Light tema** (`@theme` blogu):
- `--color-primary` / `--color-primary-dark`: Mavi (#2563eb / #1d4ed8)
- `--color-diff-insert-bg` / `--color-diff-insert-text`: Pastel yesil (rgba(0,128,0,0.08) / #27ae60)
- `--color-diff-delete-bg` / `--color-diff-delete-text`: Pastel kirmizi (rgba(255,0,0,0.08) / #c0392b)
- `--color-surface` / `--color-surface-secondary`: Beyaz / Acik gri (#f9fafb)
- `--color-border`: Gri kenarlik (#e5e7eb)
- `--color-text-primary` / `--color-text-secondary`: Koyu / Acik metin
- `--color-accent-blue` / `--color-accent-green` / `--color-accent-gray`: Review karti accent cizgi renkleri
**Dark tema** (`.dark body` CSS override ile ayni degiskenleri gecersiz kilar):
- `--color-surface`: #1a1a2e, `--color-surface-secondary`: #16213e
- `--color-border`: #2d3748, `--color-text-primary`: #e2e8f0, `--color-text-secondary`: #718096
**Dark mode gecisi**: `<html>` elemanina `.dark` class eklenir/cikarilir, tercih `localStorage('katip-theme')` ile saklanir.
Ozel siniflar: `.review-card`, `.review-card-accent-*`, `.review-card-actions` (hover-to-reveal), `.review-action-btn`, `.review-separator`

## Gereksinimler (Platform Bazli)
| Platform | Gereksinimler |
|----------|---------------|
| Tumu | Go 1.25+, Node.js 18+, Wails v3 CLI |
| Windows | Windows 10+ (WebView2, Win11'de dahili) |
| macOS | macOS 10.15+, Xcode Command Line Tools (`xcode-select --install`) |
| Linux (Debian/Ubuntu) | `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`, `gcc` |
| Linux (Fedora) | `gtk3-devel`, `webkit2gtk4.1-devel`, `gcc` |
| Linux (Arch) | `gtk3`, `webkit2gtk-4.1`, `gcc` |

## Gelistirme Komutlari
```bash
cd katip
# Gelistirme modu (hot reload) -- tum platformlar
wails3 dev
# Binding'leri yeniden olustur (Go API degisince)
wails3 generate bindings
# Frontend build
cd frontend && npm run build
```

### Windows
```powershell
# Zamanlama sorunu varsa: once vite baslat, sonra Go binary calistir
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
# .app bundle olustur
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
# AppImage olustur
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
# Docker imajini kur (ilk seferde)
wails3 task setup:docker
# Herhangi bir platformdan hedef platforma derle
wails3 task darwin:build ARCH=arm64
wails3 task linux:build ARCH=amd64
# Tam paket (mevcut platform)
wails3 build
```

## Dosya Depolama Konumlari
Tum veriler `os.UserConfigDir()/Katip/` altinda saklanir:
| Veri | Alt Dizin |
|------|-----------|
| Uygulama ayarlari | `config.json` |
| llama-server binary + kutuphaneler | `llama-server/` (Win: .exe+.dll, macOS: binary+.dylib+.metal, Linux: binary+.so) |
| llama-server indirme arsivi | `llama-server/*.zip` (tekrar acmak icin saklanir) |
| GGUF model dosyalari | `models/<model>.gguf` |
| Indirme gecici dosyalari | `models/<model>.gguf.part` |

**Platform bazli kok dizin:**
| Platform | `os.UserConfigDir()` Sonucu |
|----------|----------------------------|
| Windows | `C:\Users\<user>\AppData\Roaming\Katip\` |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

## Cross-Platform Notlar
- **Subprocess sonlandirma**: macOS/Linux'ta SIGTERM (graceful), Windows'ta Process.Kill() -- `signal_unix.go` / `signal_windows.go`
- **WebView**: Windows → WebView2 (Edge), macOS → WebKit (dahili), Linux → WebKitGTK 4.1
- **llama-server asset secimi**: `downloader.go` → `findAssetName()` fonksiyonu `runtime.GOOS` + `runtime.GOARCH` ile platforma uygun GitHub release asset'ini secer
- **Arsiv cikarma**: `.exe`, `.dll`, `.so`, `.dylib`, `.metal`, `LICENSE` dosyalari tutulur; binary'ye `GetLlamaServerPath()` ile erisilir (Windows'ta `.exe` uzantisi otomatik eklenir)
- **macOS build**: CGO gerekli, `MACOSX_DEPLOYMENT_TARGET=10.15`. Universal binary (Intel+ARM) `lipo` ile olusturulur
- **Linux build**: CGO gerekli, `libgtk-3-dev` + `libwebkit2gtk-4.1-dev` bagimliliklari
- **Windows build**: `CGO_ENABLED=0`, syso yoksa `-ldflags "-H windowsgui"` gerekli
- **Docker cross-compile**: `build/docker/Dockerfile.cross` ile herhangi bir platformdan darwin/linux/windows hedeflerine derleme

## Bilinen Sinirlamalar ve Yeni Epikler
- Dosya acma/kaydetme henüz yok. Gelecekte tek paket `.kitap` standardina tasinacak.
- RAG Motoru: Olay örgusu tutarliligi AI uzerinden denetlenecek.
- Veri izleme `.md` tabanlı Git Diff emulatorune gecirilecek.
- Streaming token destegi henuz yok.
- Birden fazla paragraf eszamanli iyilestirme henuz desteklenmiyor.
- `wails3 dev` bazen zamanlama sorunu yasayabilir (vite dev server Go binary'den sonra hazir olamaz)
- Windows'ta syso dosyasi yoksa `go build` icin `-ldflags "-H windowsgui"` gereklidir.
