# Katip — LLM Bağlam Özeti

> Bu dosya **LLM.md** — AI asistanlar ve otomatik kod araçları için token-verimli proje özeti. Geliştirici kurulum: `README.md`. Mimari derinlik: `ARCHITECTURE.md`. Stil kuralları: `DESIGN.md`.

## Kimlik

Katip: çevrimdışı, yerel LLM destekli Türkçe masaüstü metin düzenleyici. Go 1.25 / Wails v3 / Svelte 5 / Tailwind 4 / TipTap 2.11. Cross-platform (Win/macOS/Linux).

## Epik hedefler (kısmen kodlanmış)

- `.kitap`: ZIP paket — metin, meta, medya, font, diff geçmişi, RAG özet paneli (backend API mevcut)
- Git-diff geçmişi: Kalıcı İsim-Tarih satır log (oturum içi versionHistoryStore mevcut)
- RAG: Tutarlılık taraması + özet paneli (CheckConsistency mevcut)
- `.tarz`: Yayınevi jargon profilleri (profiles/*.tarz, CheckTarzStyle mevcut)

## Stack

| Katman | Teknoloji |
|--------|-----------|
| Backend | Go, Wails v3 alpha.74 |
| Frontend | Svelte 5 runes, TypeScript |
| Editor | TipTap / ProseMirror |
| LLM | llama-server subprocess, `/v1/chat/completions` (SSE streaming) |
| Diff | sergi/go-diff (kelime), PM Decoration (çoklu reviewId) |
| Spell | hunspell-wasm + tdd-ai/hunspell-tr |
| Config | `os.UserConfigDir()/Katip/config.json` |
| Models | `Katip/models/*.gguf` |
| Server | `Katip/llama-server/` |
| Tarz | `Katip/profiles/*.tarz` |

## Dizin (kritik)

```
main.go
internal/service/katip.go       # 26 public Wails method
internal/llm/{client,manager,downloader,models,signal_*}.go
internal/tarz/tarz.go
internal/diff/engine.go
frontend/src/App.svelte
frontend/src/app.css
frontend/src/lib/components/     # Editor, Toolbar, Review*, Toast, ...
frontend/src/lib/editor/         # diffDecorations, aiReviewService, spellcheck*
frontend/src/lib/stores/*.svelte.ts  # $state — .svelte.ts ZORUNLU
frontend/bindings/               # wails3 generate bindings (gitignore)
frontend/public/dictionaries/tr_TR.{aff,dic}
```

## KatipService API (26)

`Greet`, `ImproveParagraph(id,text,plotSummary,mode)`, `ImproveSelection(id,text,plotSummary,mode)`, `GeneratePlotSummary(fullText)`, `CheckConsistency(text,plotSummary)`, `CancelImprovement`, `GetImproveStreamProgress`, `GetTarzProfiles`, `CheckTarzStyle(profile,text)`, `GetLLMStatus`, `GetServerLog`, `GetConfig`, `UpdateConfig`, `StartLLMServer`, `StopLLMServer`, `CheckSetupStatus`, `CheckLlamaServer`, `DownloadLlamaServer`, `ReextractLlamaServer`, `GetDownloadProgress`, `GetModelCatalog`, `GetInstalledModels`, `DownloadModel`, `GetModelDownloadProgress`, `SaveKitap`, `LoadKitap`

## Veri modelleri

```go
DiffResult{ParagraphID, Summary, Original, Improved, Diffs[], ChangeType, RuleName}
DiffItem{Type: "equal"|"insert"|"delete", Text}
Review{paragraphPos, mode, kind, status: pending|accepted|rejected|partial}
AppConfig{ModelPath, ServerBinary, ServerHost, ServerPort, CtxSize, Threads, SystemPrompt}
```

## AI iyileştirme akışı

`aiReviewService.requestImprovement` → `ImproveParagraph|ImproveSelection(plotSummary, mode)` → system prompt (bağlam+mod) + user `<DÜZELT>` → SSE stream → `ComputeWordDiff` → `reviewStore.addReview(paragraphPos)` → `applyDiffForReview(reviewId)` → ReviewPanel (granüler ✓/✕) → `acceptReview` / `rejectReview`

## LLM prompt

Sistem: düzeltme motoru + mod kuralları + plotSummary bağlamı. Kullanıcı: yalnızca `<DÜZELT>...</DÜZELT>`. temp=0.15, top_p=0.9, stream=true.

## Bilinen sınırlamalar

Dosya aç/kaydet UI yok (.kitap API var). Paralel paragraf iyileştirme yok. Kalıcı versiyon geçmişi yok. Onayda biçimlendirme kaybı. Sözlük ekleme oturum bazlı.

## Komutlar

```bash
wails3 dev
wails3 generate bindings
cd frontend && npm run build
```

## Tasarım (özet)

Word işlevsellik + Notion minimalizm. Toast bildirimleri, mod dropdown, review toplu aksiyonlar. Detay: `DESIGN.md`.
