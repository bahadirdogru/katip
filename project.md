# Katip - Profesyonel Türkçe Metin Düzenleyici

## 1. Vizyon

CPU üzerinde yerel LLM ile çalışan, Word tarzı "Track Changes" fonksiyonelliği (kelime bazlı diff, editör içi inline markup, onayla/reddet) ile Notion tarzı minimal tasarım sunan masaüstü Türkçe metin düzenleyici. Windows, macOS (Intel & Apple Silicon) ve Linux destekler. Tamamen çevrimdışı çalışır -- kullanıcı verileri bilgisayardan çıkmaz.

---

## 2. Temel Teknoloji Kararları

### A. Inference Stratejisi: llama-server Subprocess

BitNet.cpp yerine **llama.cpp'nin llama-server** bileşeni kullanılır.

- **Neden**: BitNet.cpp, llama.cpp'nin eski ve bakımsız bir fork'u. CGO entegrasyonu özellikle Windows'ta karmaşık ve kırılgan.
- **Nasıl**: Go, `llama-server` binary'sini subprocess olarak başlatır. Tüm iletişim OpenAI-uyumlu HTTP API (`/v1/chat/completions`) üzerinden yapılır.
- **Avantaj**: CGO karmaşıklığı sıfır, herhangi bir GGUF model dosyasıyla çalışır, hata ayıklama kolay.
- **BitNet Desteği**: llama.cpp, TQ1_0/TQ2_0 ternary formatlarını native destekler. İleride Türkçe 1.58-bit model çıktığında sıfır kod değişikliğiyle çalışır.

### B. Model Stratejisi: Model-Agnostik Esnek Mimari + Yerleşik Katalog

Mimari tamamen model-agnostik. Kullanıcı ayarlar panelinden herhangi bir GGUF model dosyasını seçebilir. Bunun yanında yerleşik bir model kataloğu sunulur ve modeller uygulama içinden tek tıkla HuggingFace'den indirilir (resume desteği ile).

**Yerleşik Model Kataloğu:**

| ID | Model | Boyut | Min RAM | Dil | Varsayılan |
|----|-------|-------|---------|-----|------------|
| `turkcell-7b-q4km` | Turkcell-LLM-7b-v1 Q4_K_M | ~4.5 GB | 8 GB | Türkçe | Evet |
| `openr1-qwen-7b-tr-q4km` | OpenR1-Qwen-7B-Turkish Q4_K_M | ~4.5 GB | 8 GB | Türkçe | - |
| `qwen25-3b-q4km` | Qwen2.5-3B-Instruct Q4_K_M | ~2.0 GB | 4 GB | Çok dilli | - |
| `bitnet-2b-4t` | BitNet b1.58-2B-4T | ~1.1 GB | 2 GB | İngilizce | - |

### C. Cross-Platform Otomatik Kurulum Sistemi

İlk açılışta Setup Wizard `os.UserConfigDir()/Katip/` dizinini tarar, eksik bileşenleri tespit eder ve adım adım yönlendirir. Daha önce indirilmiş dosyalar otomatik algılanır, config doldurulur ve wizard atlanır.

**Platform bazlı yapılandırma dizini** (`os.UserConfigDir()` + `Katip/`):

| Platform | Dizin |
|----------|-------|
| Windows | `%AppData%\Katip\` (`C:\Users\<user>\AppData\Roaming\Katip\`) |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

- **llama-server**: GitHub Releases API üzerinden platforma uygun sürüm otomatik indirilir. `downloader.go` içindeki `findAssetName()` fonksiyonu `runtime.GOOS` ve `runtime.GOARCH` ile doğru asset'i seçer:
  - Windows: `win-cpu-x64` / `win-cpu-arm64` → `.exe` + `.dll` dosyaları
  - macOS: `mac-arm64` / `mac-x64` → binary + `.dylib` + `.metal` dosyaları
  - Linux: `ubuntu-x64` / `ubuntu-arm64` → binary + `.so` dosyaları
- **GGUF Model**: HuggingFace'den indirme (`.part` dosyasına yazar, resume desteği: Range header). Platform bağımsız.
- **Subprocess Yönetimi**: macOS/Linux'ta SIGTERM ile graceful shutdown, Windows'ta Process.Kill() (`signal_unix.go` / `signal_windows.go`)

### D. Akıllı Hata Tespiti

llama-server subprocess çöktüğünde log analizi yapılır:
- `"failed to allocate"` → BELLEK_YETERSIZ (RAM uyarısı + küçük model önerisi)
- `"not a valid gguf"` → MODEL_BOZUK
- `"address already in use"` → PORT_KULLANILIYOR

---

## 3. Teknoloji Yığını

| Katman | Teknoloji | Detay |
|--------|-----------|-------|
| Backend | Go 1.25+ / Wails v3 (alpha.74) | Masaüstü uygulama çatısı |
| Frontend | Svelte 5 (runes) + TypeScript | `$state`, `$derived`, `$props`, `$effect` |
| Editör | TipTap 2.11 (ProseMirror çekirdeği) | Zengin metin düzenleme |
| Stil | Tailwind CSS 4 | `@tailwindcss/vite` plugin, class tabanlı dark mode |
| Tasarım | Word fonksiyonelliği + Notion minimalizmi | hover-to-reveal, pastel track changes |
| LLM | llama-server subprocess | OpenAI-uyumlu HTTP API |
| Diff | `sergi/go-diff` (Go, kelime bazlı) | ProseMirror Decoration (frontend inline markup) |
| Yazım Denetimi | hunspell-wasm (WebAssembly) | tdd-ai/hunspell-tr Türkçe sözlükleri |
| Config | JSON dosyası | `os.UserConfigDir()/Katip/config.json` |

---

## 4. Uygulanan Özellikler

- **İlk Açılış Kurulum Sihirbazı**: Eksik bileşen tespiti, adım adım yönlendirme, otomatik algılama
- **Word Tarzı Track Changes**: Kelime bazlı diff, editör içi inline markup (silinen kırmızı üstü çizili, eklenen yeşil altı çizili), onayla/reddet
- **Gece/Gündüz Modu**: Tailwind dark mode, tek tıkla tema değişimi, tercih `localStorage` ile hatırlanır
- **Canlı Durum Işığı**: Header'da kırmızı/sarı/yeşil ışık ile AI sunucu durumu (5 sn polling)
- **Notion Tarzı Tasarım**: Minimal header, hover-to-reveal butonlar, pastel renkler, temiz tipografi
- **AI Metin İyileştirme**: Toolbar'daki mavi "AI İyileştir" butonu veya paragraf üzerine gelince hover butonu
- **Türkçe Yazım Denetleyicisi**: hunspell-wasm ile çevrimdışı, kırmızı dalgalı alt çizgi, sağ tık ile öneri popup'ı
- **Akıllı Hata Tespiti**: RAM yetersizliği, model hatası gibi sorunlar renkli banner'larla gösterilir
- **Zengin Metin Editörü**: Başlıklar, kalın, italik, listeler, alıntı blokları
- **Otomatik Kurulum**: llama-server ve GGUF modelleri uygulama içinden tek tıkla indirme
- **Model Kataloğu**: Önceden tanımlı Türkçe modeller (varsayılan badge), harici GGUF desteği
- **Özelleştirilebilir AI Prompt**: Sistem prompt'unu ayarlar panelinden düzenleme
- **Cross-Platform**: Windows, macOS (Intel & Apple Silicon) ve Linux desteği. Platform-spesifik llama-server indirme, graceful shutdown, native paketleme (NSIS/MSIX, .app bundle, AppImage/deb/rpm)

---

## 5. Kullanıcı Deneyimi ve Onay Akışı

### AI İyileştirme Akışı

1. Kullanıcı toolbar'daki mavi **AI İyileştir** butonuna veya paragraf hover butonuna tıklar
2. Backend: `ImproveParagraph(id, text)` → llama-server'a HTTP POST (`<DÜZELT>` etiketi, temperature=0.15)
3. Yanıt: `cleanLLMOutput()` ile etiket/önek temizlenir → `ComputeWordDiff()` ile kelime bazlı diff
4. Frontend: `reviewStore.addReview(result)` + `applyDiffDecorations()` ile editör içi inline markup
5. ReviewPanel'de Notion tarzı kart gösterilir (hover-to-reveal butonlar)
6. **Onayla**: `clearDecorations()` + `tr.replaceWith()` → metin güncellenir
7. **Reddet**: `clearDecorations()` + `reviewStore.rejectReview()` → orijinal korunur

### İlk Açılış Setup Wizard

1. `CheckSetupStatus()`: llama-server var mı? zip var mı? model var mı? .part var mı?
2. Config boşsa ama dosyalar mevcutsa → otomatik config doldur
3. Status: `"ready"` → wizard atlanır | diğer → wizard başlar
4. Adımlar: Hoşgeldiniz → llama-server kurulumu → Model indirme → Tamamlandı

### Türkçe Yazım Denetimi

1. `initSpellChecker()`: hunspell-wasm + `tr_TR.aff` / `tr_TR.dic` yüklenir
2. ProseMirror Plugin: Her doc değişikliğinde debounce (400ms) → `buildSpellDecorations()`
3. Sağ tık → `SpellSuggestion.svelte` popup → öneri listesi veya "Sözlüğe ekle"

---

## 6. Wails Binding API (KatipService - 17 Method)

| Method | Parametre | Dönüş | Açıklama |
|--------|-----------|-------|----------|
| `Greet(name)` | string | string | Test metodu |
| `ImproveParagraph(id, text)` | string, string | DiffResult | Paragrafı AI ile iyileştir |
| `GetLLMStatus()` | - | map | running, healthy, endpoint, modelPath, lastError |
| `GetServerLog()` | - | string | llama-server stdout/stderr logu |
| `GetConfig()` | - | AppConfig | Uygulama ayarları |
| `UpdateConfig(cfg)` | AppConfig | error | Ayarları güncelle ve kaydet |
| `StartLLMServer()` | - | error | llama-server subprocess başlat |
| `StopLLMServer()` | - | error | llama-server'ı durdur |
| `CheckSetupStatus()` | - | map | İlk açılış durumu |
| `CheckLlamaServer()` | - | map | installed, path, zipExists |
| `DownloadLlamaServer()` | - | error | GitHub'dan llama-server indir |
| `ReextractLlamaServer()` | - | error | Mevcut zip'ten yeniden çıkar |
| `GetDownloadProgress()` | - | DownloadProgress | llama-server indirme durumu |
| `GetModelCatalog()` | - | []ModelInfo | Mevcut GGUF model listesi |
| `GetInstalledModels()` | - | []string | İndirilmiş model ID'leri |
| `DownloadModel(modelID)` | string | error | HuggingFace'den model indir |
| `GetModelDownloadProgress()` | - | DownloadProgress | Model indirme durumu |

---

## 7. Proje Dosya Yapısı

```
katip/
├── main.go                              # Wails app entry point, embed frontend/dist
├── go.mod                               # module katip, Go 1.25
├── internal/
│   ├── service/katip.go                 # KatipService: Wails'e expose edilen ana API (17 method)
│   ├── llm/
│   │   ├── client.go                    # llama-server HTTP client, <DÜZELT> etiketli prompt, cleanLLMOutput
│   │   ├── manager.go                   # llama-server subprocess yaşam döngüsü + analyzeServerLog
│   │   ├── downloader.go               # GitHub Releases API: llama-server indirme + zip açma (cross-platform)
│   │   ├── models.go                   # GGUF model kataloğu + HuggingFace indirme (resume desteği)
│   │   ├── signal_unix.go              # Unix (macOS/Linux): SIGTERM ile graceful shutdown
│   │   └── signal_windows.go           # Windows: Process.Kill() ile sonlandırma
│   └── diff/engine.go                   # go-diff wrapper: ComputeDiff, ComputeWordDiff
├── frontend/
│   ├── src/
│   │   ├── App.svelte                   # Ana layout + Setup Wizard + durum ışığı polling + dark mode
│   │   ├── main.ts                      # Svelte 5 mount() entry point
│   │   ├── app.css                      # Tailwind tema (light/dark) + track changes + spell check
│   │   └── lib/
│   │       ├── components/
│   │       │   ├── Editor.svelte        # TipTap editör, hover AI butonu, spell check entegrasyonu
│   │       │   ├── Toolbar.svelte       # Biçimlendirme + mavi AI İyileştir butonu + applyDiffDecorations
│   │       │   ├── ReviewPanel.svelte   # Sağ panel: Notion tarzı header + separator
│   │       │   ├── ReviewCard.svelte    # hover-to-reveal, accent çizgi, kırpılmış diff
│   │       │   ├── SetupWizard.svelte   # İlk açılış kurulum sihirbazı (4 adım)
│   │       │   ├── SpellSuggestion.svelte # Yazım önerisi popup'ı (sağ tık menüsü)
│   │       │   └── SettingsDialog.svelte # Kurulum + model katalogu + hata banner'ları
│   │       ├── editor/
│   │       │   ├── diffDecorations.ts   # ProseMirror Plugin: inline delete + widget insert
│   │       │   ├── spellChecker.ts      # hunspell-wasm wrapper: init, testSpelling, getSuggestions
│   │       │   ├── spellcheckPlugin.ts  # ProseMirror Plugin: wavy underline (debounced)
│   │       │   └── extensions.ts        # TipTap extension re-exports
│   │       └── stores/
│   │           └── reviewStore.svelte.ts # Svelte 5 $state class: Review CRUD (.svelte.ts zorunlu)
│   ├── public/dictionaries/             # Türkçe hunspell sözlük dosyaları
│   │   ├── tr_TR.aff                    # Affix kuralları (~2.2 MB)
│   │   └── tr_TR.dic                    # Sözlük (~34.5 MB)
│   ├── bindings/                        # Wails otomatik oluşturur (gitignore'da)
│   └── package.json
├── claude.md                            # AI asistan bağlam dosyası
├── project.md                           # Proje mimari dokümanı
└── README.md                            # Kullanıcı dokümanı
```

---

## 8. Donanım ve Kaynak Planlaması

| Bileşen | Kaynak Kullanımı | Strateji |
|---------|-----------------|----------|
| Model (7B Q4_K_M) | 4 - 5 GB RAM | Standart GGUF quantization |
| Model (3B Q4_K_M) | 1.5 - 2 GB RAM | Hafif alternatif |
| Model (2B 1.58-bit) | 0.4 GB RAM | Native ternary (Türkçe zayıf) |
| Wails + Svelte UI | 150 - 300 MB | Windows: WebView2, macOS: WebKit, Linux: WebKitGTK |
| llama-server | 50 - 100 MB | Subprocess overhead |
| İşlemci (i5) | %20 - %60 (yük altında) | llama.cpp SIMD optimizasyonları |

---

## 9. Veri Depolama

Tüm veriler `os.UserConfigDir()/Katip/` altında saklanır (platform bağımsız):

| Veri | Alt Dizin |
|------|-----------|
| Uygulama ayarları | `config.json` |
| llama-server binary + kütüphaneler | `llama-server/` (.exe+.dll / binary+.dylib+.metal / binary+.so) |
| llama-server indirme arşivi | `llama-server/*.zip` |
| GGUF model dosyaları | `models/<model>.gguf` |
| İndirme geçici dosyaları | `models/<model>.gguf.part` |

**Platform bazlı tam yollar:**

| Platform | Kök Dizin |
|----------|-----------|
| Windows | `C:\Users\<user>\AppData\Roaming\Katip\` |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

---

## 10. Bilinen Sınırlamalar ve Gelecek Çalışmalar

### Mevcut Sınırlamalar
- Dosya açma/kaydetme henüz yok
- Streaming token desteği henüz yok (şu an tam yanıt bekleniyor)
- Birden fazla paragraf eşzamanlı iyileştirme henüz desteklenmiyor
- Sözlüğe eklenen kelimeler oturum bazlıdır (kalıcı değil)
- `wails3 dev` bazen zamanlama sorunu yaşayabilir

### Gelecek Çalışmalar
- Dosya açma/kaydetme/dışa aktarma
- Streaming token desteği (kısmi yanıt gösterimi)
- Çoklu paragraf eşzamanlı iyileştirme
- Kalıcı kullanıcı sözlüğü
- Türkçe'ye özel native 1.58-bit model desteği (çıktığında)

---

## 11. Kritik Kütüphane ve Kaynak Linkleri

1. **Inference Runtime:** [llama.cpp / llama-server](https://github.com/ggml-org/llama.cpp)
2. **App Framework:** [Wails v3](https://v3.wails.io/)
3. **Editor Engine:** [TipTap](https://tiptap.dev/)
4. **Diff Logic:** [sergi/go-diff](https://github.com/sergi/go-diff)
5. **UI Framework:** [Svelte 5](https://svelte.dev/)
6. **CSS Framework:** [Tailwind CSS 4](https://tailwindcss.com/)
7. **Yazım Denetimi:** [hunspell-wasm](https://www.npmjs.com/package/hunspell-wasm)
8. **Türkçe Sözlük:** [tdd-ai/hunspell-tr](https://github.com/tdd-ai/hunspell-tr)
