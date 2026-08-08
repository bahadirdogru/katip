# Katip — Teknik Mimari

> Bu dosya **ARCHITECTURE.md** — sistemin teknik mimarisi, teknoloji kararları, veri akışları, API yüzeyi ve epik planları. GitHub giriş ve kurulum: `README.md`. AI asistan özeti: `LLM.md`. UI/stil: `DESIGN.md`.

---

## 1. Vizyon ve felsefe

Katip, CPU üzerinde yerel LLM ile çalışan, Git benzeri versiyon kontrol hedefi taşıyan ve Notion tarzı minimal arayüz sunan masaüstü Türkçe metin düzenleyicidir. Tamamen çevrimdışı; kullanıcı verileri cihazdan çıkmaz.

Temel odak karmaşık takım/yetki yönetimi **değil**; bireysel yazar ve editörün ekran performansını artırmak ve işbirliğini e-posta tabanlı asenkron `.kitap` paket transferiyle kolaylaştırmak.

Geleneksel Word "Track Changes" okuma hızını düşürür; dağınık klasör yapıları medya/font/meta taşımayı zorlaştırır. Katip bu sorunları `.kitap` paketi, temiz diff geçmişi ve yerel RAG hedefiyle ele alır.

---

## 2. Teknoloji kararları

### 2.1 Inference: llama-server subprocess

BitNet.cpp yerine **llama.cpp llama-server** kullanılır.

| Neden | Detay |
|-------|-------|
| BitNet.cpp | Eski fork, Windows CGO entegrasyonu kırılgan |
| llama-server | Subprocess, OpenAI-uyumlu HTTP, CGO sıfır |
| BitNet desteği | llama.cpp TQ1_0/TQ2_0 native; Türkçe model çıktığında sıfır kod değişikliği |

Go, `llama-server` binary'sini subprocess başlatır. İletişim: `POST /v1/chat/completions`.

### 2.2 Model stratejisi

Model-agnostik: ayarlardan herhangi bir GGUF seçilebilir. Yerleşik katalog HuggingFace'den tek tık indirme (`.part` resume).

| ID | Model | Boyut | Min RAM | Dil | Varsayılan |
|----|-------|-------|---------|-----|------------|
| `turkcell-7b-q4km` | Turkcell-LLM-7b-v1 Q4_K_M | ~4.5 GB | 8 GB | Türkçe | Evet |
| `openr1-qwen-7b-tr-q4km` | OpenR1-Qwen-7B-Turkish Q4_K_M | ~4.5 GB | 8 GB | Türkçe | — |
| `qwen25-3b-q4km` | Qwen2.5-3B-Instruct Q4_K_M | ~2.0 GB | 4 GB | Çok dilli | — |
| `bitnet-2b-4t` | BitNet b1.58-2B-4T | ~1.1 GB | 2 GB | İngilizce | — |

### 2.3 Cross-platform kurulum

İlk açılış Setup Wizard `os.UserConfigDir()/Katip/` tarar. Eksik bileşenleri adım adım kurar; mevcut dosyalar varsa config otomatik doldurulur.

**Platform dizinleri:**

| Platform | Kök |
|----------|-----|
| Windows | `%AppData%\Katip\` |
| macOS | `~/Library/Application Support/Katip/` |
| Linux | `~/.config/Katip/` |

**llama-server indirme** (`downloader.go` → `findAssetName()`):

| OS | Asset | Çıkarılan |
|----|-------|-----------|
| Windows | win-cpu-x64 / win-cpu-arm64 | .exe + .dll |
| macOS | mac-arm64 / mac-x64 | binary + .dylib + .metal |
| Linux | ubuntu-x64 / ubuntu-arm64 | binary + .so |

**Subprocess sonlandırma:** macOS/Linux SIGTERM (`signal_unix.go`); Windows Process.Kill (`signal_windows.go`).

### 2.4 Akıllı hata tespiti

`manager.go` → `analyzeServerLog()` log kalıpları:

| Kalıp | Tip | UI |
|-------|-----|-----|
| `failed to allocate` | BELLEK_YETERSIZ | RAM uyarısı, küçük model önerisi |
| `not a valid gguf` | MODEL_BOZUK | Kırmızı banner |
| `address already in use` | PORT_KULLANILIYOR | Port çakışması banner |

---

## 3. Teknoloji yığını

| Katman | Teknoloji | Not |
|--------|-----------|-----|
| Backend | Go 1.25+, Wails v3 (alpha.74) | Masaüstü çatı |
| Frontend | Svelte 5 runes + TypeScript | `$state`, `$derived`, `$props`, `$effect` |
| Editör | TipTap 2.11 (ProseMirror) | Zengin metin |
| Stil | Tailwind CSS 4 | `@tailwindcss/vite`, class dark mode |
| LLM | llama-server | OpenAI-uyumlu HTTP |
| Diff | sergi/go-diff + PM Decoration | Kelime bazlı backend, inline frontend |
| Yazım | hunspell-wasm + hunspell-tr | WebAssembly, debounce 400ms |
| Config | JSON | `config.json` |
| Paket (hedef) | `.kitap` | ZIP/JSON/MD bundle |

---

## 4. Epik planları ve mimari eğilimler

### Faz 1: `.kitap` offline paket

- **Kapsüllenmiş mimari**: Metin, yorumlar, meta, görseller, fontlar — tek ZIP, e-posta transferi
- **Chaptering**: Büyük eserlerde aktif bölüm belleğe; context window tasarrufu
- **Statik özet paneli**: Olay/karakter tablosu paket içinde; RAG referans belleği

### Faz 2: İzleme ve UX

- **Git kalitesinde geçmiş**: Satır/cümle "İsim-Tarih" log; merge conflict hariç
- **Asenkron yorumlar**: Çıktıya yansımayan tartışma balonları (CommentPanel mevcut)
- **Melez arayüz**: Toolbar + öğretici tooltip'ler, zoom, dark mode

### Faz 3: LLM ve yayınevi kuralları

- **RAG optimizasyonu**: Özet paneli statik referans; karakter/olay tutarlılığı
- **`.tarz` profilleri**: Jargon ve kelime kuralları import/export
- **Seri tarama**: Bölüm taraması → Öneri Kartları (Toolbar "Tümünü Tara" mevcut)

---

## 5. Mevcut özellikler (kodlanmış)

| Özellik | Durum | Not |
|---------|-------|-----|
| Setup Wizard | Mevcut | 4 adım, otomatik algılama |
| Kelime diff + inline markup | Mevcut | Git-diff'e evrilecek |
| Gece/gündüz modu | Mevcut | localStorage |
| AI durum ışığı | Mevcut | 5 sn polling |
| Notion tarzı review panel | Mevcut | hover-to-reveal |
| AI iyileştirme | Mevcut | Toolbar + hover |
| Türkçe yazım denetimi | Mevcut | hunspell-wasm |
| Akıllı hata banner'ları | Mevcut | SettingsDialog |
| Zengin metin editörü | Mevcut | Başlık, liste, alıntı |
| Otomatik kurulum | Mevcut | llama-server + model |
| Model kataloğu | Mevcut | Varsayılan badge |
| Özelleştirilebilir prompt | Mevcut | Ayarlar |
| Yorum paneli | Mevcut | CommentPanel |
| Hikaye özeti paneli | Mevcut | SummaryPanel (RAG hedefi) |
| Toplu tarama | Mevcut | Sıralı, durdur/devam, özet toast |
| Çoklu diff dekorasyon | Mevcut | reviewId bazlı PM plugin |
| Granüler diff onay/red | Mevcut | Kart içi ✓/✕ per değişiklik |
| İyileştirme modları | Mevcut | fix/shorten/flow/formal |
| Seçim bazlı iyileştirme | Mevcut | ImproveSelection API |
| Streaming yanıt + iptal | Mevcut | SSE + CancelImprovement |
| Toast bildirimleri | Mevcut | toastStore + Toast.svelte |
| LLM guard | Mevcut | llmState → buton disable |
| Tutarlılık taraması | Mevcut | CheckConsistency + özet paneli |
| `.tarz` profilleri | Mevcut | profiles/*.tarz, CheckTarzStyle |
| Versiyon geçmişi | Mevcut | Oturum içi versionHistoryStore |
| aiReviewService | Mevcut | Merkezi AI review orchestration |
| Cross-platform | Mevcut | Win/macOS/Linux paketleme |

---

## 6. Sistem mimarisi

```
┌─────────────────────────────────────────────────────────┐
│                      Wails v3 Host                       │
│  ┌──────────────────┐      ┌─────────────────────────┐ │
│  │   Go Backend     │      │   Svelte 5 Frontend     │ │
│  │                  │      │                         │ │
│  │  KatipService    │◄────►│  App.svelte             │ │
│  │  LLM Client      │ bind │  Editor (TipTap)        │ │
│  │  LLM Manager     │      │  Toolbar                │ │
│  │  Diff Engine     │      │  ReviewPanel            │ │
│  │  Downloader      │      │  CommentPanel           │ │
│  │  Model Catalog   │      │  SummaryPanel           │ │
│  │                  │      │  SetupWizard            │ │
│  │                  │      │  SettingsDialog         │ │
│  │                  │      │  SpellChecker (WASM)    │ │
│  └────────┬─────────┘      └─────────────────────────┘ │
│           │ HTTP POST /v1/chat/completions              │
│           ▼                                             │
│  ┌──────────────────┐                                   │
│  │  llama-server    │  subprocess                       │
│  │  + GGUF model    │                                   │
│  └──────────────────┘                                   │
└─────────────────────────────────────────────────────────┘
```

### WebView katmanı

| Platform | Runtime |
|----------|---------|
| Windows | WebView2 (Edge) |
| macOS | WebKit |
| Linux | WebKitGTK 4.1 |

---

## 7. Veri akışları

### 7.1 AI iyileştirme

```
Kullanıcı [AI İyileştir / hover / Seçimi İyileştir]
  → aiReviewService.requestImprovement(editor, {scope, mode, llmState})
  → Wails: ImproveParagraph(id, text, plotSummary, mode) | ImproveSelection(...)
  → Go: llmClient.ImproveWithOptions(text, plotSummary, mode)
  → System prompt: düzeltme kuralları + mod + plotSummary (bağlam)
  → User mesaj: yalnızca <DÜZELT>metin</DÜZELT>
  → HTTP POST llama-server (stream:true, SSE chunks)
  → cleanLLMOutput → ComputeWordDiff → DiffResult
  → reviewStore.addReview (paragraphPos, mode, changeType)
  → applyDiffForReview(reviewId) → çoklu PM dekorasyon
  → ReviewPanel kart → granüler veya tam onay/red
  → acceptReview: pos bazlı replace + versionHistoryStore
```

### 7.2 Toplu tarama

`collectParagraphs()` ile tüm textblock'lar toplanır (min 10 karakter). Sırayla `ImproveParagraph`; çoklu diff dekorasyon; durdur/devam; tamamlanınca toast özeti.

### 7.3 llama-server yaşam döngüsü

```
StartLLMServer()
  → subprocess başlat (config: binary, model, port, ctx, threads)
  → health check polling
GetLLMStatus() → {running, healthy, endpoint, modelPath, lastError}
StopLLMServer() → platform-specific signal/kill
Çökme → analyzeServerLog() → errorType → SettingsDialog banner
```

### 7.4 llama-server indirme

```
DownloadLlamaServer()
  → goroutine: GitHub Releases API → son sürüm
  → findAssetName(GOOS, GOARCH)
  → HTTP GET zip → extract → Katip/llama-server/
  → config.ServerBinary güncelle
Frontend: GetDownloadProgress() polling
ReextractLlamaServer(): mevcut zip'ten yeniden çıkar
```

### 7.5 Model indirme

```
DownloadModel(modelID)
  → goroutine: HuggingFace resolve URL
  → .part dosyasına yaz (Range resume)
  → tamamlanınca .gguf rename
  → config.ModelPath güncelle
Frontend: GetModelDownloadProgress() polling
```

### 7.6 Setup Wizard

```
onMount → CheckSetupStatus()
  → llama-server? zip? model? .part?
  → config boş + dosyalar var → auto-fill + save
  → status: ready | zip_found | llama_missing | model_partial | model_missing
  → ready → ana editör | diğer → SetupWizard
Adımlar: Hoşgeldiniz → llama → model → Tamamlandı
```

### 7.7 Header durum ışığı

```
pollLLMStatus() her 5sn
  healthy → yeşil "AI Hazır"
  running → sarı ping "Yükleniyor"
  else → kırmızı "AI Kapalı"
Tıklama → SettingsDialog
```

### 7.8 Yazım denetimi

```
initSpellChecker() → fetch tr_TR.aff/dic → hunspell-wasm
spellcheckPlugin: doc change debounce 400ms
  → buildSpellDecorations → tokenize → testSpelling
  → Decoration.inline("spell-error") dalgalı kırmızı
Sağ tık → SpellSuggestion popup → öneri / sözlüğe ekle
```

### 7.9 LLM iletişim detayları

- **Sistem prompt**: Metin düzeltme motoru + mod ek kuralları + plotSummary bağlamı (ayrı bölüm)
- **Kullanıcı mesajı**: yalnızca `<DUZELT>metin</DUZELT>` — bağlam system prompt'ta
- **Modlar**: `fix` | `shorten` | `flow` | `formal`
- **Parametreler**: temperature=0.15, top_p=0.9, stream=true
- **İptal**: `CancelImprovement()` → client.cancelled flag
- **cleanLLMOutput**: Etiket, önek ("Düzeltilmiş metin:"), gereksiz tırnak temizliği

---

## 8. Wails binding sistemi

Go public method'lar otomatik JS binding:

```bash
wails3 generate bindings  # → frontend/bindings/ (gitignore)
```

Import örneği:

```typescript
import { ImproveParagraph } from '../bindings/katip/internal/service/katipservice.js';
```

### KatipService API (26 method)

| Method | Parametre | Dönüş | Açıklama |
|--------|-----------|-------|----------|
| `Greet(name)` | string | string | Test |
| `ImproveParagraph(id, text, plotSummary, mode)` | string×4 | DiffResult | AI paragraf iyileştirme |
| `ImproveSelection(id, text, plotSummary, mode)` | string×4 | DiffResult | Seçim bazlı iyileştirme |
| `GeneratePlotSummary(fullText)` | string | string | Otomatik hikaye özeti |
| `CheckConsistency(text, plotSummary)` | string×2 | string | Tutarlılık denetimi |
| `CancelImprovement()` | — | — | Aktif LLM isteğini iptal |
| `GetImproveStreamProgress()` | — | map | Streaming ilerleme |
| `GetTarzProfiles()` | — | []string | .tarz profil listesi |
| `CheckTarzStyle(profile, text)` | string×2 | []StyleViolation | Tarz kural denetimi |
| `GetLLMStatus()` | — | map | running, healthy, endpoint, ... |
| `GetServerLog()` | — | string | stdout/stderr log |
| `GetConfig()` | — | AppConfig | Ayarlar |
| `UpdateConfig(cfg)` | AppConfig | error | Kaydet |
| `StartLLMServer()` | — | error | Subprocess başlat |
| `StopLLMServer()` | — | error | Durdur |
| `CheckSetupStatus()` | — | map | Kurulum durumu |
| `CheckLlamaServer()` | — | map | installed, path, zipExists |
| `DownloadLlamaServer()` | — | error | GitHub indirme |
| `ReextractLlamaServer()` | — | error | Zip yeniden çıkar |
| `GetDownloadProgress()` | — | DownloadProgress | İndirme durumu |
| `GetModelCatalog()` | — | []ModelInfo | Model listesi |
| `GetInstalledModels()` | — | []string | İndirilmiş ID'ler |
| `DownloadModel(modelID)` | string | error | HuggingFace indirme |
| `GetModelDownloadProgress()` | — | DownloadProgress | Model indirme |
| `SaveKitap(path, doc)` | string, Document | error | .kitap kaydet |
| `LoadKitap(path)` | string | Document | .kitap yükle |

### Review modeli (frontend)

```typescript
interface Review {
  id: string;
  paragraphId: string;
  paragraphPos: number;
  selectionFrom?: number;
  selectionTo?: number;
  summary: string;
  original: string;
  improved: string;
  diffs: DiffItem[];
  status: 'pending' | 'accepted' | 'rejected' | 'partial';
  mode?: 'fix' | 'shorten' | 'flow' | 'formal';
  kind?: 'improvement' | 'consistency' | 'style';
  changeType?: 'spelling' | 'grammar' | 'style' | 'consistency';
}
```

### Veri modelleri

```go
type DiffResult struct {
    ParagraphID string     `json:"paragraphId"`
    Summary     string     `json:"summary"`
    Original    string     `json:"original"`
    Improved    string     `json:"improved"`
    Diffs       []DiffItem `json:"diffs"`
}

type DiffItem struct {
    Type string `json:"type"` // equal | insert | delete
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

---

## 9. Proje dosya yapısı

```
katip/
├── main.go
├── go.mod
├── internal/
│   ├── service/katip.go
│   ├── llm/
│   │   ├── client.go
│   │   ├── manager.go
│   │   ├── downloader.go
│   │   ├── models.go
│   │   ├── signal_unix.go
│   │   └── signal_windows.go
│   ├── tarz/tarz.go
│   └── diff/engine.go
├── frontend/
│   ├── src/
│   │   ├── App.svelte
│   │   ├── main.ts
│   │   ├── app.css
│   │   └── lib/
│   │       ├── components/
│   │       │   ├── Editor.svelte
│   │       │   ├── Toolbar.svelte
│   │       │   ├── ReviewPanel.svelte
│   │       │   ├── ReviewCard.svelte
│   │       │   ├── CommentPanel.svelte
│   │       │   ├── SummaryPanel.svelte
│   │       │   ├── SetupWizard.svelte
│   │       │   ├── SpellSuggestion.svelte
│   │       │   ├── Toast.svelte
│   │       │   └── SettingsDialog.svelte
│   │       ├── editor/
│   │       │   ├── diffDecorations.ts
│   │       │   ├── aiReviewService.ts
│   │       │   ├── spellChecker.ts
│   │       │   ├── spellcheckPlugin.ts
│   │       │   └── extensions.ts
│   │       └── stores/
│   │           ├── reviewStore.svelte.ts
│   │           ├── toastStore.svelte.ts
│   │           ├── versionHistoryStore.svelte.ts
│   │           ├── commentStore.svelte.ts
│   │           ├── documentStore.svelte.ts
│   │           └── settingsStore.svelte.ts
│   ├── public/dictionaries/tr_TR.{aff,dic}
│   ├── bindings/          # otomatik, gitignore
│   └── package.json
├── README.md
├── ARCHITECTURE.md
├── LLM.md
└── DESIGN.md
```

---

## 10. Frontend mimarisi

### Svelte 5 kuralları

- Durum: `let x = $state(value)`
- Türetilmiş: `$derived`
- Efekt: `$effect`
- Props: `$props()`
- Event: `onclick` (Svelte 4 `on:click` kullanılmaz)
- Mount: `mount(App, { target })`
- **`$` önek yasağı**: `$from` → `const resolved = state.selection.$from`
- **`.svelte.ts` zorunlu**: `$state` içeren store dosyaları

### Store'lar

| Store | Sorumluluk |
|-------|------------|
| `reviewStore` | AI düzeltme kartları, granüler diff, paragraphPos |
| `toastStore` | Kullanıcı bildirimleri |
| `versionHistoryStore` | Onaylanan değişiklik geçmişi (oturum) |
| `commentStore` | Yorum thread'leri |
| `documentStore` | plotSummary, belge meta |
| `settingsStore` | fontSize, fontFamily |

### Editör eklentileri

- `diffDecorations.ts`: PM Plugin, çoklu reviewId dekorasyon, reviewId bazlı clear
- `aiReviewService.ts`: Merkezi improve/accept/reject/scroll orchestration
- `spellcheckPlugin.ts`: debounced wavy underline
- TipTap extensions: bold, italic, underline, strike, highlight, headings, lists, blockquote, text align, color

### Layout

```
Header (logo, AI ışık, panel sekmeleri, tema, ayarlar)
Toolbar (zoom, biçimlendirme, AI İyileştir, Tümünü Tara)
Main (max-w-3xl editör alanı, px-20 py-8)
Right aside (w-72): Düzeltmeler | Yorumlar | Özet
Footer (kelime/karakter sayacı)
```

---

## 11. Donanım ve kaynak planlaması

| Bileşen | Kaynak | Strateji |
|---------|--------|----------|
| Model 7B Q4_K_M | 4–5 GB RAM | Varsayılan Türkçe |
| Model 3B Q4_K_M | 1.5–2 GB RAM | Hafif alternatif |
| Model 2B 1.58-bit | ~0.4 GB RAM | Türkçe zayıf |
| Wails + Svelte UI | 150–300 MB | WebView |
| llama-server | 50–100 MB | Subprocess |
| CPU (i5) | %20–60 yük altında | llama.cpp SIMD |

---

## 12. Veri depolama

| Veri | Alt dizin |
|------|-----------|
| Ayarlar | `config.json` |
| llama-server | `llama-server/` |
| Zip arşivi | `llama-server/*.zip` |
| GGUF modeller | `models/<model>.gguf` |
| İndirme geçici | `models/<model>.gguf.part` |
| `.tarz` profilleri | `profiles/*.tarz` |

---

## 13. Cross-platform derleme

| Platform | CGO | Not |
|----------|-----|-----|
| Windows | 0 | `-ldflags "-H windowsgui"` syso yoksa |
| macOS | 1 | `MACOSX_DEPLOYMENT_TARGET=10.15`, universal `lipo` |
| Linux | 1 | gtk3 + webkit2gtk-4.1-dev |
| Docker | cross | `build/docker/Dockerfile.cross` |

---

## 14. Bilinen sınırlamalar

### Teknik sınırlamalar

- Dosya aç/kaydet UI henüz yok (`.kitap` backend mevcut, SaveKitap/LoadKitap API var)
- Paralel çoklu paragraf iyileştirme yok (llama-server tek istek; sıralı tarama)
- Kalıcı sözlük ekleme oturum bazlı
- Kalıcı versiyon geçmişi yok (oturum içi versionHistoryStore)
- `wails3 dev` zamanlama sorunu olabilir

### UX sınırlamaları

- Onayda zengin metin biçimlendirmesi (bold, link) plain text'e dönüşür
- Tutarlılık taraması ilk 5 paragrafla sınırlı (performans)
- `.tarz` denetimi basit string eşleştirme (regex değil)

### Kodlanacak epikler

- `.kitap` tam UI entegrasyonu
- Kalıcı Git-benzeri cümle diff log
- Gelişmiş RAG motoru (statik özet kartları)
- `.tarz` profil editörü UI

### Kapsam dışı (iptal)

- IDML, DOCX, EPUB çıktı (salt Markdown hedefi)
- Merge conflict / eşzamanlı takım senkronizasyonu

---

## 15. Kaynak linkleri

1. [llama.cpp / llama-server](https://github.com/ggml-org/llama.cpp)
2. [Wails v3](https://v3.wails.io/)
3. [TipTap](https://tiptap.dev/)
4. [sergi/go-diff](https://github.com/sergi/go-diff)
5. [Svelte 5](https://svelte.dev/)
6. [Tailwind CSS 4](https://tailwindcss.com/)
7. [hunspell-wasm](https://www.npmjs.com/package/hunspell-wasm)
8. [tdd-ai/hunspell-tr](https://github.com/tdd-ai/hunspell-tr)
