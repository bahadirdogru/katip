# Katip

> Bu dosya **README.md** — GitHub vitrini ve geliştirici giriş noktasıdır. Projeye katılmak, derlemek ve çalıştırmak için buradan başlayın. Teknik mimari için `ARCHITECTURE.md`, AI bağlam özeti için `LLM.md`, stil ve tasarım kuralları için `DESIGN.md` dosyalarına bakın.

Yerel AI destekli, çevrimdışı Türkçe metin düzenleyici. Word tarzı izleme esnekliği, Git kalitesinde versiyon geçmişi hedefi ve Notion tarzı minimal arayüz bir arada. Windows, macOS (Intel & Apple Silicon) ve Linux.

Veriler cihazdan çıkmaz. llama-server ve GGUF modeller uygulama içinden indirilebilir; hunspell-wasm ile Türkçe yazım denetimi tamamen yerel çalışır.

## Dokümantasyon

| Dosya | Amaç |
|-------|------|
| **README.md** (bu dosya) | GitHub vitrini, kurulum, derleme, geliştirici rehberi |
| **ARCHITECTURE.md** | Teknik mimari, veri akışları, API, epik planları |
| **LLM.md** | AI asistanlar için token-verimli proje özeti |
| **DESIGN.md** | UI/UX, renkler, tipografi, bileşen stilleri |

## Özellikler (özet)

- **AI metin iyileştirme**: Paragraf ve seçim bazlı düzeltme, 4 mod (Düzelt/Kısalt/Akıcılaştır/Resmileştir), kelime diff, granüler onay/red
- **Toplu tarama**: Belgedeki tüm paragrafları sırayla AI ile denetleme; durdur/devam, özet bildirimi
- **Review paneli**: Tümünü onayla/reddet, kart navigasyonu, değişiklik bloğu bazında ✓/✕
- **Streaming yanıt**: llama-server SSE akışı; işlem sırasında iptal butonu
- **Hikaye özeti (RAG)**: Manuel/otomatik bağlam; tutarlılık taraması
- **`.tarz` profilleri**: Yayınevi jargon kuralları denetimi (`profiles/*.tarz`)
- **Versiyon geçmişi**: Onaylanan AI değişiklikleri oturum içi log
- **Türkçe yazım denetimi**: hunspell-wasm, sağ tık önerileri + "AI ile düzelt"
- **Toast bildirimleri**: Hata, boş sonuç ve tarama özeti geri bildirimi
- **Klavye kısayolları**: `Ctrl+Shift+I` iyileştir, `Ctrl+Enter` onayla, `Esc` reddet
- **Kurulum sihirbazı**, **model kataloğu**, **gece/gündüz modu**, **cross-platform**

Gelecek epikler: `.kitap` paket I/O, kalıcı Git-diff geçmişi. Ayrıntılar `ARCHITECTURE.md` içinde.

## Gereksinimler

Tüm platformlar:

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails v3 CLI](https://v3.wails.io/)

```bash
go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
```

| Platform | Ek gereksinimler |
|----------|------------------|
| Windows 10+ | WebView2 (Win11'de dahili) |
| macOS 10.15+ | Xcode Command Line Tools (`xcode-select --install`) |
| Debian/Ubuntu | `libgtk-3-dev libwebkit2gtk-4.1-dev gcc` |
| Fedora | `gtk3-devel webkit2gtk4.1-devel gcc` |
| Arch | `gtk3 webkit2gtk-4.1 gcc` |

## Hızlı başlangıç

```bash
git clone <repo-url> katip
cd katip
cd frontend && npm install && cd ..
wails3 dev
```

Go API değiştiğinde binding'leri yeniden üretin:

```bash
wails3 generate bindings
```

### Zamanlama sorunu (`wails3 dev`)

Vite hazır olmadan Go binary açılırsa frontend boş kalabilir. Önce Vite, sonra binary:

**Windows:**

```powershell
cd frontend; npx vite --port 9245 --strictPort
# Ayrı terminalde:
cd C:\Git\katip
go build -ldflags "-H windowsgui" -o bin/katip.exe .
$env:FRONTEND_DEVSERVER_URL="http://localhost:9245"; .\bin\katip.exe
```

**macOS / Linux:**

```bash
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -o bin/katip .
FRONTEND_DEVSERVER_URL="http://localhost:9245" ./bin/katip
```

## İlk kullanım

1. Setup Wizard eksik bileşenleri tespit eder (llama-server, varsayılan model)
2. Önceden indirilmiş dosyalar varsa config otomatik doldurulur, wizard atlanır
3. Ayarlardan AI sunucusunu başlatın (header'daki durum ışığı yeşile döner)
4. Metin yazın; toolbar'dan **AI İyileştir** (mod seçin), **Seçimi İyileştir** veya **Tümünü Tara** kullanın
5. Özet panelinden hikaye bağlamı ekleyin; **Tutarlılık Tara** ve **Tarz Denetimi** ile gelişmiş denetim yapın

## Mimari (kısa)

```
Wails v3
├── Go Backend (KatipService, LLM manager, diff engine, downloader)
└── Svelte 5 Frontend (TipTap, review panel, spell check WASM)
         ↕ HTTP (OpenAI uyumlu)
    llama-server subprocess + GGUF model
```

Detaylı mimari, veri akışları ve API tablosu: `ARCHITECTURE.md`.

## Proje yapısı

```
katip/
├── main.go
├── internal/
│   ├── service/katip.go      # Wails API (26 method)
│   ├── llm/                  # client, manager, downloader, models
│   ├── tarz/                 # .tarz profil dosyaları
│   └── diff/engine.go
├── frontend/src/
│   ├── App.svelte
│   ├── app.css
│   └── lib/
│       ├── components/       # Editor, Toolbar, ReviewPanel, Toast, ...
│       ├── editor/           # diffDecorations, aiReviewService, spellcheck
│       └── stores/           # reviewStore, toastStore, versionHistoryStore
├── README.md
├── ARCHITECTURE.md
├── LLM.md
└── DESIGN.md
```

## Veri depolama

Tüm yapılandırma ve model verileri `os.UserConfigDir()/Katip/` altında:

| Veri | Konum |
|------|-------|
| Ayarlar | `config.json` |
| llama-server | `llama-server/` |
| GGUF modeller | `models/*.gguf` |
| İndirme geçici | `models/*.gguf.part` |
| `.tarz` profilleri | `profiles/*.tarz` |

| Platform | Kök dizin |
|----------|-----------|
| Windows | `%AppData%\Katip\` |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

## Derleme

```bash
wails3 build                              # Üretim (mevcut platform)
wails3 task darwin:build:universal        # macOS universal
wails3 task darwin:package                # macOS .app
wails3 task linux:create:appimage
wails3 task linux:create:deb
wails3 task linux:create:rpm
wails3 task setup:docker                  # Cross-compile (ilk sefer)
wails3 task darwin:build ARCH=arm64
wails3 task linux:build ARCH=amd64
```

Windows'ta syso yoksa: `go build -ldflags "-H windowsgui" -o bin/katip.exe .`

## Geliştirme notları

- **Svelte 5 runes** zorunlu; `$state` kullanan store dosyaları `.svelte.ts` uzantılı olmalı
- **Binding'ler** `frontend/bindings/` altında; gitignore'da, Go API değişince yeniden üretin
- **Stil değişiklikleri** `DESIGN.md` kurallarına uygun yapılmalı (`app.css` + Tailwind 4)
- **UI metinleri** Türkçe; kod ve commit mesajları projedeki mevcut convention'ı izleyin

## Model kataloğu

| Model | Boyut | Min RAM | Dil | Varsayılan |
|-------|-------|---------|-----|------------|
| Turkcell-LLM-7b-v1 | ~4.5 GB | 8 GB | Türkçe | Evet |
| OpenR1-Qwen-7B-Turkish | ~4.5 GB | 8 GB | Türkçe | — |
| Qwen2.5-3B-Instruct | ~2.0 GB | 4 GB | Çok dilli | — |
| BitNet b1.58-2B-4T | ~1.1 GB | 2 GB | İngilizce | — |

Harici GGUF dosyası ayarlar panelinden seçilebilir.

## Teknolojiler

- [Go](https://go.dev/) + [Wails v3](https://v3.wails.io/)
- [Svelte 5](https://svelte.dev/) + TypeScript
- [TipTap](https://tiptap.dev/) / ProseMirror
- [Tailwind CSS 4](https://tailwindcss.com/)
- [llama.cpp](https://github.com/ggml-org/llama.cpp) (llama-server)
- [go-diff](https://github.com/sergi/go-diff)
- [hunspell-wasm](https://www.npmjs.com/package/hunspell-wasm) + [hunspell-tr](https://github.com/tdd-ai/hunspell-tr)

## Lisans

Depo kökündeki LICENSE dosyasına bakın (varsa).
