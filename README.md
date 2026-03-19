# Katip

Yerel AI destekli profesyonel Turkce metin duzenleyici. Word tarzi Track Changes fonksiyonelligi ile Notion tarzi minimal tasarim.

Katip tamamen cevrimdisi calisir -- verileriniz bilgisayarinizdan cikmaz.

Windows, macOS (Intel & Apple Silicon) ve Linux destekler.

## Ozellikler

- **Ilk Acilis Kurulum Sihirbazi**: Uygulama ilk acildiginda yapilandirma dizinini tarar, eksik bilesenleri tespit eder ve adim adim yonlendirir. Zaten indirilmis dosyalari otomatik algilar, config'i doldurur ve wizard'i atlar.
- **Word Tarzi Track Changes**: Kelime bazli diff, editor ici inline markup (silinen kirmizi ustu cizili, eklenen yesil alti cizili), onayla/reddet
- **Gece/Gunduz Modu**: Tailwind dark mode ile tek tikla tema degisimi, tercih hatirlanir
- **Canli Durum Isigi**: Header'da kirmizi/sari/yesil isik ile AI sunucu durumu aninda gorunur
- **Notion Tarzi Tasarim**: Minimal header, hover-to-reveal butonlar, pastel renkler, temiz tipografi
- **AI Metin Iyilestirme**: Toolbar'daki belirgin mavi "AI Iyilestir" butonuna tiklayin veya paragraf uzerine gelin
- **Turkce Yazim Denetleyicisi**: hunspell-wasm (WebAssembly) + tdd-ai/hunspell-tr sozlukleri ile cevrimdisi yazim kontrolu. Yanlis yazilan kelimeler kirmizi dalgali alt cizgi ile isaretlenir, sag tik ile oneri popup'i acilir.
- **Akilli Hata Tespiti**: RAM yetersizligi, model hatasi gibi sorunlar belirgin uyari banner'lari ile gosterilir
- **Zengin Metin Editoru**: Basliklar, kalin, italik, listeler, alinti bloklari
- **Otomatik Kurulum**: llama-server ve GGUF modelleri uygulama icinden tek tikla indirin
- **Model Secimi**: Onceden tanimli Turkce model katalogundan secim yapin veya kendi GGUF modelinizi kullanin. Varsayilan model (Turkcell-LLM-7b-v1) badge ile isaretlidir.
- **Ozellestirilebilir AI Prompt**: Sistem prompt'unu ayarlar panelinden duzenleyin
- **Cross-Platform**: Windows, macOS (Intel & Apple Silicon) ve Linux uzerinde calisir

## Gereksinimler

Tum platformlar icin ortak:

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails v3 CLI](https://v3.wails.io/)

```bash
go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Windows

- Windows 10+ (WebView2 gerekli, Windows 11'de dahili)

### macOS

- macOS 10.15 (Catalina) veya uzeri
- Xcode Command Line Tools

```bash
xcode-select --install
```

### Linux

- GTK 3 ve WebKitGTK 4.1 kutuphaneleri
- GCC veya Clang derleyicisi

```bash
# Debian / Ubuntu
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev gcc

# Fedora
sudo dnf install gtk3-devel webkit2gtk4.1-devel gcc

# Arch Linux
sudo pacman -S gtk3 webkit2gtk-4.1 gcc
```

## Hizli Baslangic

```bash
cd katip

# Bagimliliklari yukle
cd frontend && npm install && cd ..

# Gelistirme modunda calistir (tum platformlar)
wails3 dev
```

### Platforma ozel notlar

**Windows** - Zamanlama sorunu yasanirsa:

```powershell
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -ldflags "-H windowsgui" -o bin/katip.exe .
$env:FRONTEND_DEVSERVER_URL="http://localhost:9245"; .\bin\katip.exe
```

**macOS / Linux** - Zamanlama sorunu yasanirsa:

```bash
cd frontend && npx vite --port 9245 --strictPort &
cd .. && go build -o bin/katip .
FRONTEND_DEVSERVER_URL="http://localhost:9245" ./bin/katip
```

Uygulama ilk acildiginda:

1. **Setup Wizard** otomatik olarak baslar ve eksik bilesenleri tespit eder
2. llama-server ve varsayilan model (Turkcell-LLM-7b-v1, ~4.5 GB) adim adim indirilir
3. Daha once indirilmis dosyalar varsa otomatik algilanir ve wizard atlanir
4. Kurulum tamamlaninca ana ekrana gecilir

Sonraki kullanimlarda:

1. Sag ustteki **ayarlar** butonundan AI sunucusunu **Baslat** (header'daki isik yesile donecek)
2. Editore metin yazin, toolbar'daki mavi **AI Iyilestir** butonuna tiklayin
3. Yazim hatalari otomatik olarak kirmizi dalgali alt cizgi ile isaretlenir, sag tik ile oneri alin
4. Gece/gunduz modu icin header'daki ay/gunes ikonuna tiklayin

## Model Katalogu

Uygulama icinden dogrudan indirilebilir modeller:

| Model | Boyut | Min RAM | Dil | Aciklama |
|-------|-------|---------|-----|----------|
| **Turkcell-LLM-7b-v1** (Varsayilan) | ~4.5 GB | 8 GB | Turkce | Mistral 7B tabanli, en iyi Turkce kalitesi |
| **OpenR1-Qwen-7B-Turkish** | ~4.5 GB | 8 GB | Turkce | Qwen2.5 tabanli, reasoning/dusunme yetenegi |
| **Qwen2.5-3B-Instruct** | ~2.0 GB | 4 GB | Cok dilli | Hafif ve hizli, sinirli kaynakli sistemler icin |
| **BitNet b1.58-2B-4T** | ~1.1 GB | 2 GB | Ingilizce | Ultra hizli, cok dusuk kaynak kullanimi |

RAM yetersizligi durumunda uygulama otomatik olarak uyari gosterir ve daha kucuk model onerir.

Katalog disi herhangi bir GGUF model dosyasini da ayarlar panelinden yol belirterek kullanabilirsiniz.

## Mimari

```
+-------------------------------------------------+
|                   Wails v3                      |
|  +----------------+    +---------------------+  |
|  |   Go Backend   |    | Svelte 5 Frontend   |  |
|  |                |    |                     |  |
|  | KatipService   |<-->| TipTap Editor       |  |
|  | LLM Client     |    | Inline Decorations  |  |
|  | LLM Manager    |    | ReviewPanel (Notion) |  |
|  | Diff Engine    |    | SetupWizard         |  |
|  | Downloader     |    | SpellChecker(WASM)  |  |
|  |                |    | SettingsDialog      |  |
|  +-------+--------+    +---------------------+  |
|          |                                       |
|          v                                       |
|  +----------------+                              |
|  | llama-server   |  (subprocess)                |
|  |   + GGUF       |                              |
|  +----------------+                              |
+-------------------------------------------------+
```

- **Backend (Go)**: LLM subprocess yonetimi, HTTP istemcisi (DUZELT etiketi + cleanLLMOutput), kelime bazli diff, akilli hata tespiti (RAM/model/port), ayar yonetimi, otomatik indirme, ilk acilis durum taramasi (CheckSetupStatus)
- **Frontend (Svelte 5 + TypeScript)**: TipTap zengin metin editoru, editor ici inline track changes (ProseMirror Decoration), Notion tarzi review kartlari (hover-to-reveal), Setup Wizard (ilk acilis), hunspell-wasm Turkce yazim denetleyicisi, gece/gunduz modu, canli AI durum isigi, indirme ilerleme cubuklari, hata banner'lari
- **Iletisim**: Wails binding sistemi (Go method'lari otomatik olarak JS fonksiyonlarina donusur)

## Proje Yapisi

```
katip/
+-- main.go                     # Uygulama giris noktasi
+-- internal/
|   +-- service/katip.go        # Ana API servisi (17 method)
|   +-- llm/
|   |   +-- client.go           # llama-server HTTP istemcisi + prompt yonetimi
|   |   +-- manager.go          # llama-server surec yonetimi + hata analizi
|   |   +-- downloader.go       # llama-server otomatik indirme (platform-agnostik)
|   |   +-- models.go           # GGUF model katalogu (IsDefault, MinRAM) ve indirme
|   |   +-- signal_unix.go      # Unix (macOS/Linux) SIGTERM graceful shutdown
|   |   +-- signal_windows.go   # Windows process termination
|   +-- diff/engine.go          # Metin fark hesaplama (karakter + kelime bazli)
+-- frontend/
|   +-- src/
|       +-- App.svelte           # Ana bilesen + Setup Wizard + durum isigi + dark mode
|       +-- app.css              # Tailwind tema (light/dark) + track changes + spell check stilleri
|       +-- lib/
|           +-- components/      # UI bilesenleri (Editor, Toolbar, ReviewCard, SetupWizard, SpellSuggestion, vb.)
|           +-- editor/          # ProseMirror diff plugin + spellcheck plugin + decorations
|           +-- stores/          # Durum yonetimi (reviewStore)
|   +-- public/dictionaries/    # Turkce hunspell sozluk dosyalari (tr_TR.aff, tr_TR.dic)
+-- project.md                   # Detayli mimari dokuman
```

## Veri Depolama

Tum veriler yerel sistemde, `os.UserConfigDir()` altinda saklanir:

| Veri | Windows | macOS | Linux |
|------|---------|-------|-------|
| Uygulama ayarlari | `%AppData%\Katip\config.json` | `~/Library/Application Support/Katip/config.json` | `~/.config/Katip/config.json` |
| llama-server | `%AppData%\Katip\llama-server\` | `~/Library/Application Support/Katip/llama-server/` | `~/.config/Katip/llama-server/` |
| GGUF modelleri | `%AppData%\Katip\models\` | `~/Library/Application Support/Katip/models/` | `~/.config/Katip/models/` |

## Derleme

```bash
# Uretim surumu (mevcut platform)
wails3 build

# macOS universal binary (Intel + Apple Silicon)
wails3 task darwin:build:universal

# macOS .app bundle
wails3 task darwin:package

# Linux AppImage
wails3 task linux:create:appimage

# Linux deb/rpm paketleri
wails3 task linux:create:deb
wails3 task linux:create:rpm

# Docker ile cross-compile (herhangi bir platformdan)
wails3 task setup:docker   # ilk seferde
wails3 task darwin:build ARCH=arm64
wails3 task linux:build ARCH=amd64
```

Calistirilabilir dosya `bin/` dizininde olusturulur.

## Teknolojiler

- [Go](https://go.dev/) + [Wails v3](https://v3.wails.io/) -- masaustu uygulama catisi
- [Svelte 5](https://svelte.dev/) + [TypeScript](https://www.typescriptlang.org/) -- frontend
- [TipTap](https://tiptap.dev/) (ProseMirror) -- zengin metin editoru
- [Tailwind CSS 4](https://tailwindcss.com/) -- stil
- [llama.cpp](https://github.com/ggml-org/llama.cpp) -- yerel LLM inference (MIT lisans)
- [go-diff](https://github.com/sergi/go-diff) -- metin karsilastirma
- [hunspell-wasm](https://www.npmjs.com/package/hunspell-wasm) -- WebAssembly tabanli yazim denetleyicisi
- [tdd-ai/hunspell-tr](https://github.com/tdd-ai/hunspell-tr) -- Turkce hunspell sozlukleri

## Lisans

Bu proje gelistirme asamasindadir.
