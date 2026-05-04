# Katip

> 💡 **Geliştirici Notu**: Sistemin detaylı yeni '.kitap' mimarisi ve onaylanmayan (hariç tutulan) kapsam sınırları için `project.md` dosyasına, AI ajanlarına yönelik optimize bağlam kısıtlamaları için `claude.md` dosyasına bakınız.

Yerel AI destekli profesyonel Turkce metin duzenleyici. Word tarzi izleme esnekliği ve Git kalitesinde versiyon geçmişi ile Notion tarzi minimal tasarim harmanlanarak bireysel yazar/editör odaklı üretilmiştir.

Katip tamamen cevrimdisi calisir -- verileriniz ve yazarların eserleri bilgisayarinizdan cikmaz. Çalışma dosyalarının (.kitap formatında) kişiler arası e-posta üzerinden taşınarak eşzamanlı hiyerarşik zorluklardan uzak, asenkron bir senkronizasyonla işletilmesi temeline dayanır.

Gereksinim duyulan LLM altyapısı (llama.cpp vb.) arka planda tamamen şeffaf kontrol edilir, Windows, macOS (Intel & Apple Silicon) ve Linux destekler.

## Ozellikler

- **Yeni `.kitap` Her Şey Dahil Paket Mimarisi**: Karmaşık klasör yapıları yerine; sadece metinleri değil, eklenmiş fontları, medyaları (resimler), detaylı meta verileri (kelime hedefleri vb.), versiyon tarihçesini ve onay bekleyen satıriçi yorumları (Comments) içeren tek bir ZIP bundle. Dosyayı e-posta ile attığınızda tüm donanım ve editör kurgunuz bozulmadan aktarılır.
- **Git Kalitesinde Versiyon Geçmişi**: Geleneksel Word tarzi karmaşık "Track Changes" (Değişiklikleri İzle) mimarisi yerine, kodlamacılar için ideal olan temiz "İsim-Tarih" log tabanlı, satır/cümle düzeyinde geçmiş dökümü.
- **AI Tutarlılık Motoru (RAG)**: AI, 400 sayfalık metni RAM'e yüklemek yerine; `.kitap` paketinin içine gömülü "Olay ve Karakter Özeti" referansını tarar. "3. bölümde sarışın olan karakter 7. bölümde esmer denildi" türündeki kurgu hatalarını saniyeler içinde zekice yakalar.
- **Dışa/İçe Açık `.tarz` Profilleri**: Yayınevinin kurumsal jargonunu veya özel kelime filtrelerini ("olanak" değil "imkân" olacak vb.) XML profil dosyası gibi (`.tarz`) dışarıdan programa takın. Seri taramalar ile anlatım bozukluğu kontrollerinizi bu profile göre tek tıkla yapın.
- **Ilk Acilis Kurulum Sihirbazi**: Uygulama ilk acildiginda yapilandirma dizinini tarar, eksik bilesenleri tespit eder ve adim adim yonlendirir. Zaten indirilmis dosyalari otomatik algilar, config'i doldurur ve wizard'i atlar.
- **Gece/Gunduz Modu**: Tailwind dark mode ile tek tikla tema degisimi, tercih hatirlanir. Yeni sayfayı ölçeklendirme (Zoom & Font) özellikleri.
- **Canli Durum Isigi**: Header'da kirmizi/sari/yesil isik ile AI sunucu durumu aninda gorunur
- **Melez Tasarim**: Word tarzı tepede her an erişilebilir öğretici (Tooltip'li kısayol öğreten) araç çubuğu (Toolbar) + Notion minimalizmi özellikleri bir arada sunulur.
- **Turkce Yazim Denetleyicisi**: hunspell-wasm (WebAssembly) + tdd-ai/hunspell-tr sozlukleri ile cevrimdisi yazim kontrolu. Yanlis yazilan kelimeler kirmizi dalgali alt cizgi ile isaretlenir, sag tik ile oneri popup'i acilir.
- **Akilli Hata Tespiti**: RAM yetersizligi, model hatasi gibi sorunlar belirgin uyari banner'lari ile gosterilir
- **Otomatik Kurulum**: llama-server ve GGUF modelleri uygulama icinden tek tikla indirin
- **Model Secimi**: Onceden tanimli Turkce model katalogundan secim yapin veya kendi GGUF modelinizi kullanin. Varsayilan model (Turkcell-LLM-7b-v1) badge ile isaretlidir.
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
2. Editore metin yazin, toolbar'daki mavi **AI Iyilestir** veya "Toplu Tara" (Öneri Kartları oluştur) butonlarına tiklayin
3. Yazim hatalari otomatik olarak kirmizi dalgali alt cizgi ile isaretlenir, sag tik ile oneri alin
4. Gece/gunduz modu ve sayfa zoom özellikleri için header'daki ikonlara tıklayın

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

- **Backend (Go)**: LLM subprocess yonetimi, HTTP istemcisi (DUZELT etiketi + cleanLLMOutput), cümle bazli diff (Git-history emülasyonu), akilli hata tespiti (RAM/model/port), ayar yonetimi, otomatik indirme, .kitap bundle derleyici.
- **Frontend (Svelte 5 + TypeScript)**: TipTap zengin metin editoru, Notion tarzi review kartlari (hover-to-reveal), Setup Wizard (ilk acilis), hunspell-wasm Turkce yazim denetleyicisi, Toolbar öğreten Tooltip sistemi, asenkron Yorum blokları.
- **Iletisim**: Wails binding sistemi (Go method'lari otomatik olarak JS fonksiyonlarina donusur)

## Proje Yapisi

```
katip/
+-- main.go                     # Uygulama giris noktasi
+-- internal/
|   +-- service/katip.go        # Ana API servisi
|   +-- llm/
|   |   +-- client.go           # llama-server HTTP istemcisi + prompt yonetimi
|   |   +-- manager.go          # llama-server surec yonetimi + hata analizi
|   |   +-- downloader.go       # llama-server otomatik indirme (platform-agnostik)
|   |   +-- models.go           # GGUF model katalogu (IsDefault, MinRAM) ve indirme
|   |   +-- signal_unix.go      # Unix (macOS/Linux) SIGTERM graceful shutdown
|   |   +-- signal_windows.go   # Windows process termination
|   +-- diff/engine.go          # Metin fark hesaplama (karakter/kelime bazli track history)
+-- frontend/
|   +-- src/
|       +-- App.svelte           # Ana bilesen + Setup Wizard + durum isigi + dark mode
|       +-- app.css              # Tailwind tema (light/dark) + track changes + spell check stilleri
|       +-- lib/
|           +-- components/      # UI bilesenleri (Editor, Toolbar, ReviewCard, SetupWizard vb.)
|           +-- editor/          # ProseMirror diff plugin + spellcheck plugin + decorations
|           +-- stores/          # Durum yonetimi (reviewStore.svelte.ts zincirleri)
|   +-- public/dictionaries/    # Turkce hunspell sozluk dosyalari (tr_TR.aff, tr_TR.dic)
+-- project.md                   # Detayli mimari kararlar ve epik yol haritası (Architecture)
+-- claude.md                    # AI asistan optimize kurallar paneli
```

## Veri Depolama

Tum veriler yerel sistemde, `os.UserConfigDir()` altinda saklanir:

| Veri | Windows | macOS | Linux |
|------|---------|-------|-------|
| Uygulama ayarlari | `%AppData%\Katip\config.json` | `~/Library/Application Support/Katip/config.json` | `~/.config/Katip/config.json` |
| llama-server | `%AppData%\Katip\llama-server\` | `~/Library/Application Support/Katip/llama-server/` | `~/.config/Katip/llama-server/` |
| GGUF modelleri | `%AppData%\Katip\models\` | `~/Library/Application Support/Katip/models/` | `~/.config/Katip/models/` |

*(Not: Projeler tamamen .kitap uzantısıyla e-posta aktarımında lokal bilgisayarda birleştirilir).*

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
- [Svelte 5](https://svelte.dev/) + [TypeScript](https://www.typescriptlang.org/) -- frontend (Runes API)
- [TipTap](https://tiptap.dev/) (ProseMirror) -- zengin metin editoru
- [Tailwind CSS 4](https://tailwindcss.com/) -- stil
- [llama.cpp](https://github.com/ggml-org/llama.cpp) -- yerel LLM inference motoru
- [go-diff](https://github.com/sergi/go-diff) -- metin karsilastirma / geçmiş diff çekirdeği
- [hunspell-wasm](https://www.npmjs.com/package/hunspell-wasm) -- WebAssembly tabanli yazim denetleyicisi
- [tdd-ai/hunspell-tr](https://github.com/tdd-ai/hunspell-tr) -- Turkce hunspell sozlukleri
