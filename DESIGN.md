# Katip — Tasarım ve Stil Rehberi

> Bu dosya **DESIGN.md** — UI/UX felsefesi, renk sistemi, tipografi, bileşen stilleri ve etkileşim kalıpları. Teknik mimari: `ARCHITECTURE.md`. Geliştirici kurulum: `README.md`. AI özeti: `LLM.md`.

---

## 1. Tasarım felsefesi

Katip **melez arayüz** hedefler: Word'ün işlevselliği (her zaman erişilebilir toolbar, biçimlendirme, track changes) ile Notion'un minimalizmi (temiz tipografi, hover-to-reveal, ince accent çizgiler, pastel diff renkleri).

### İlkeler

| İlke | Uygulama |
|------|----------|
| Minimal görünüm | İnce border, az padding, düşük görsel gürültü |
| İşlev öncelikli | Toolbar her zaman görünür; AI butonu belirgin mavi |
| Hover-to-reveal | Review kartı onay/red butonları sadece hover'da |
| Pastel diff | Kırmızı/yeşil track changes yüksek kontrast değil, pastel arka plan |
| Öğretici UI | Toolbar tooltip'leri klavye kısayollarını gösterir |
| Okuma odaklı editör | max-w-3xl, line-height 1.8, geniş yan boşluk (px-20) |
| Tutarlı tema | CSS değişkenleri + Tailwind 4 `@theme`; dark mode class tabanlı |

### Hedef kullanıcı deneyimi

Yazar uzun metin okurken dikkat dağıtmayan arayüz; AI ve düzeltme işlemleri yan panelde ve inline markup ile görünür ama baskın değil. Kurulum ve hata durumları açık banner/wizard ile yönlendirilir.

---

## 2. Renk sistemi

Tüm renkler `frontend/src/app.css` içinde `@theme` ve dark override ile tanımlı. Bileşenlerde `bg-surface`, `text-text-primary` gibi Tailwind token'ları kullanılır.

### Light tema (`@theme`)

| Token | Değer | Kullanım |
|-------|-------|----------|
| `--color-primary` | `#2563eb` | AI butonu, aktif sekme, link vurgusu |
| `--color-primary-dark` | `#1d4ed8` | Primary hover |
| `--color-surface` | `#ffffff` | Ana arka plan, kart |
| `--color-surface-secondary` | `#f9fafb` | Hover arka plan, toolbar grup |
| `--color-border` | `#e5e7eb` | Ayırıcılar, kenarlıklar |
| `--color-text-primary` | `#1e293b` | Ana metin |
| `--color-text-secondary` | `#94a3b8` | İkincil metin, label |
| `--color-diff-insert-bg` | `rgba(0,128,0,0.08)` | Eklenen metin arka plan |
| `--color-diff-insert-text` | `#27ae60` | Eklenen metin |
| `--color-diff-delete-bg` | `rgba(255,0,0,0.08)` | Silinen metin arka plan |
| `--color-diff-delete-text` | `#c0392b` | Silinen metin |
| `--color-accent-blue` | `#2563eb` | Review kart pending accent |
| `--color-accent-green` | `#22c55e` | Review kart accepted accent |
| `--color-accent-gray` | `#d1d5db` | Review kart rejected accent |

### Dark tema (`.dark body` override)

| Token | Değer |
|-------|-------|
| `--color-surface` | `#1a1a2e` |
| `--color-surface-secondary` | `#16213e` |
| `--color-border` | `#2d3748` |
| `--color-text-primary` | `#e2e8f0` |
| `--color-text-secondary` | `#718096` |
| `--color-diff-insert-bg` | `rgba(39,174,96,0.15)` |
| `--color-diff-delete-bg` | `rgba(192,57,43,0.15)` |
| `--color-accent-gray` | `#4a5568` |

### Durum renkleri (sabit Tailwind / inline)

| Durum | Renk | Konum |
|-------|------|-------|
| AI hazır | `bg-green-500` | Header durum ışığı |
| AI yükleniyor | `bg-amber-400` + `animate-ping` | Header |
| AI kapalı | `bg-red-400` | Header |
| Yazım hatası | `#e53e3e` wavy underline | `.spell-error` |
| AI işleniyor | `amber-100` / `dark:amber-900/40` | AI İyileştir butonu |
| Toplu tarama | `blue-100` / `dark:blue-900/40` | Tümünü Tara butonu |
| Onay hover | `#16a34a` + `rgba(22,163,74,0.1)` | Review accept btn |
| Red hover | `#dc2626` + `rgba(220,38,38,0.1)` | Review reject btn |

### Hata banner'ları (SettingsDialog)

RAM yetersiz, model bozuk, port çakışması için renkli banner'lar; errorType'a göre arka plan ve metin rengi değişir (kırmızı/amber tonları).

---

## 3. Dark mode

**Mekanizma:** `document.documentElement.classList.toggle('dark')`

**Kalıcılık:** `localStorage.setItem('katip-theme', 'dark'|'light')`

**CSS:** `@custom-variant dark (&:where(.dark, .dark *))` — Tailwind 4 class tabanlı dark mode.

**Toggle:** Header'da ay/güneş ikonu (`App.svelte` → `toggleTheme()`).

Yeni bileşenler dark uyumlu token kullanmalı; sabit `#fff` / `#000` yerine `var(--color-*)` veya `bg-surface` gibi tema token'ları tercih edilir.

---

## 4. Tipografi

### Sistem font yığını

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
```

`-webkit-font-smoothing: antialiased` body'de aktif.

### Editör metni (ProseMirror)

| Öğe | Boyut | Ağırlık | Not |
|-----|-------|---------|-----|
| Body | `settingsStore.fontSize` px (zoom) | normal | line-height 1.8 |
| h1 | 1.875rem | 700 | letter-spacing -0.02em, mt 2rem |
| h2 | 1.5rem | 600 | letter-spacing -0.01em, mt 1.5rem |
| h3 | 1.25rem | 600 | mt 1.25rem |
| Paragraf | inherit | — | mb 0.25rem, padding 3px 2px |
| Placeholder | text-secondary | — | `data-placeholder` pseudo |

Font ailesi `settingsStore.fontFamily` ile değiştirilebilir (footer zoom A-/A+ ile boyut).

### UI metin ölçekleri

| Bağlam | Boyut | Örnek |
|--------|-------|-------|
| Header logo | `text-sm font-semibold` | "Katip" |
| Durum ışığı label | `text-[11px]` | "AI Hazır" |
| Panel sekmeleri | `text-xs` | Düzeltmeler, Yorumlar |
| Review kart diff | `text-[13px]` | Kart içi metin |
| Review meta | `text-xs` / `text-[10px]` | Özet, zaman |
| Panel header | `0.75rem uppercase` | `.review-panel-header` |
| Footer sayaç | `text-[10px]` | Kelime/karakter |
| Toolbar buton | `text-sm` | Biçimlendirme |
| Tooltip | `text-xs` | Kısayol balonu |
| Spell popup | `0.8125rem` | Öneri listesi |

---

## 5. Layout ve spacing

### Ana düzen

```
┌──────────────────────────────────────────────────────────┐
│ Header: px-4 py-1.5, border-b, shrink-0, draggable      │
├──────────────────────────────────────────────────────────┤
│ Toolbar: px-3 py-2, border-b, overflow-x-auto            │
├───────────────────────────────┬──────────────────────────┤
│ Main: flex-1, overflow-y-auto │ Aside: w-72, border-l    │
│ max-w-3xl mx-auto px-20 py-8  │ shrink-0                 │
├───────────────────────────────┴──────────────────────────┤
│ Footer: px-4 py-1.5, border-t, text-[10px]               │
└──────────────────────────────────────────────────────────┘
```

### Spacing kuralları

- Header/toolbar/footer: `shrink-0` — scroll alanı sadece main/aside
- Editör içerik: `max-w-3xl` (~768px) okuma genişliği
- Yan panel: sabit `w-72` (288px)
- Toolbar grupları: `w-px h-4 bg-border mx-1` dikey ayırıcı
- Review kart: `padding 12px 14px`, accent sol `2px` çizgi

---

## 6. Bileşen stilleri

### Header (`App.svelte`)

- Logo: `text-sm font-semibold tracking-tight`
- Durum butonu: `px-2 py-0.5 rounded text-[11px] hover:bg-surface-secondary`
- Panel sekmeleri: gruplu `border border-border bg-surface-secondary rounded p-0.5`
- Aktif sekme: `bg-surface shadow text-primary font-medium`
- Pasif sekme: `text-text-secondary hover:text-text-primary`
- Tema/ayarlar: `text-xs px-2 py-1 rounded hover:bg-surface-secondary`

### Toolbar (`Toolbar.svelte`)

**Zoom kontrolü:** `bg-surface-secondary border rounded`, A-/A+ butonları, ortada mono font boyut.

**Biçimlendirme butonları:**

- Varsayılan: `text-text-secondary hover:bg-surface-secondary`
- Aktif: `text-primary bg-primary/15`
- Boyut: `px-2.5 py-1.5 min-w-[32px] shadow-sm`

**Tooltip (öğretici):**

- Konum: `bottom-full`, `opacity-0 invisible` → `group-hover:opacity-100`
- Arka plan: `bg-gray-800 text-white`
- İçerik: label + mono kısayol (`text-[9px] text-gray-400`)
- Ok: CSS border triangle

**AI İyileştir (birincil CTA):**

- Normal: `text-white bg-primary hover:bg-primary-dark shadow-sm hover:shadow`
- İşleniyor: `text-amber-700 bg-amber-100 dark:amber-300 dark:bg-amber-900/40`
- `px-4 py-1.5 font-medium rounded-md`

**Tümünü Tara (ikincil):**

- Normal: `text-text-secondary bg-surface-secondary border border-border`
- Tarama: `text-blue-700 bg-blue-100 dark:blue-300 dark:bg-blue-900/40`

**Mod dropdown:**

- `text-xs`, `border-border`, `bg-surface-secondary`
- Seçenekler: Düzelt, Kısalt, Akıcılaştır, Resmileştir

**Seçimi İyileştir / İptal:**

- İkincil buton stili; İptal: `text-red-600`

### Toast bildirimleri (`Toast.svelte`)

- Konum: `fixed bottom-4 left-1/2`
- Tipler: info (surface), success (yeşil), warning (amber), error (kırmızı)
- Animasyon: `slideUp 0.2s ease`
- Örnek mesajlar: "AI sunucusu çalışmıyor. Ayarlardan başlatın.", "Bu metinde düzeltme önerilmedi.", "12 paragraftan 3'ünde öneri bulundu."

### Review panel (Notion tarzı)

**Header** (`.review-panel-header`):

- `0.75rem`, weight 500, uppercase, letter-spacing 0.05em, text-secondary

**Kart** (`.review-card`):

- `padding 12px 14px`, `border-radius 6px`
- Hover: `background surface-secondary`, transition 150ms
- Sol accent çizgi (`.review-card-accent`): 2px, top/bottom 8px inset

**Accent durumları:**

| Sınıf | Renk | Anlam |
|-------|------|-------|
| `.review-card-accent-pending` | accent-blue | Bekleyen |
| `.review-card-accent-accepted` | accent-green | Onaylandı |
| `.review-card-accent-rejected` | accent-gray | Reddedildi |
| `.review-card-accent-warning` | `#dc2626` | Tutarlılık uyarısı |
| `.review-card-accent-style` | `#f59e0b` | Tarz kuralı |

**Toplu aksiyonlar** (header altı):

- Tümünü Onayla: yeşil tonlu `text-[10px]` buton
- Tümünü Reddet / navigasyon ← →: secondary buton
- Tamamlananları Temizle: underline link

**Granüler diff butonları** (`.review-diff-btn`):

- 16×16px, hover'da görünür (kart hover veya focus)
- `.review-diff-btn-accept` / `.review-diff-btn-reject`

**Aktif kart:** `ring-1 ring-primary/30 bg-primary/5`

**Paragraf highlight:** `.review-highlight-pulse` — 1.5s mavi flash

- `.review-card-actions`: `opacity 0` → kart hover `opacity 1`
- `.review-action-btn`: 26×26px, rounded 4px, transparent bg
- Accept hover: yeşil ton
- Reject hover: kırmızı ton

**Ayırıcı:** `.review-separator` — 1px border, margin 0 14px

### Track changes (inline diff)

**Silinen** (`.diff-delete`):

- Arka plan: diff-delete-bg
- Renk: diff-delete-text
- `line-through`, border-radius 2px

**Eklenen** (`.diff-insert`, `.diff-insert-widget`):

- Arka plan: diff-insert-bg
- Renk: diff-insert-text
- `underline`, underline-offset 2px

Editör içi ve review kartında aynı sınıflar kullanılır — görsel tutarlılık.

### Yazım denetimi

**Hata işareti** (`.spell-error`):

- `underline wavy #e53e3e`
- `text-underline-offset 3px`
- `cursor: context-menu`

**Popup** (`.spell-popup`):

- `z-index 50`, min-width 160px, max 240px
- `border-radius 8px`, gölge (light/dark farklı yoğunluk)
- Öğe: `padding 5px 12px`, hover surface-secondary
- Ayırıcı: 1px border
- Boş durum: italic, text-secondary

### Setup Wizard

Tam ekran merkezli akış; adım göstergesi, progress bar, birincil/ikincil butonlar proje primary/secondary token'larına uyumlu. Hata mesajları kırmızı tonlu metin kutusu.

### Settings Dialog

Modal overlay; form alanları standart border/surface. Model kartları: boyut badge, varsayılan işareti, indirme progress. Sunucu log expandable bölüm. Hata banner'ları üstte renkli.

---

## 7. Etkileşim ve animasyon

| Etkileşim | Süre / tip |
|-----------|------------|
| Kart hover bg | 150ms ease |
| Review aksiyon opacity | 150ms ease |
| Action btn hover | 120ms ease |
| Spell popup item | 100ms ease |
| AI durum ping | Tailwind `animate-ping` |
| AI işleniyor | `animate-spin` emoji |
| Tema geçişi | Anında (CSS değişken swap) |

**Focus:** Toolbar butonları `focus:outline-none`; erişilebilirlik iyileştirmesi gelecekte `focus-visible` ring eklenebilir.

**Disabled:** `opacity-30` veya `opacity-40`, pointer events kapalı.

---

## 8. İkon ve emoji kullanımı

Projede metin tabanlı ikonlar ve emoji tercih edilir (harici icon kütüphanesi yok):

| Öğe | İkon |
|-----|------|
| Gece modu | 🌙 |
| Gündüz modu | ☀️ |
| Ayarlar | ⚙ |
| AI İyileştir | ✨ |
| Seçimi İyileştir | (metin buton) |
| İptal | (metin, kırmızı) |
| Tümünü Tara | 🔎 |
| İşleniyor | ⏳ (spin) |
| Geri al / yinele | ↩ ↪ |
| Onayla / reddet | ✓ ✕ |

Toolbar biçimlendirme: B, I, U, S, A ve unicode hizalama sembolleri.

---

## 9. Tailwind 4 kullanımı

- Giriş: `@import "tailwindcss"` (`app.css`)
- Plugin: `@tailwindcss/vite` (Vite config)
- Tema token'ları: `@theme { --color-* }` → `bg-primary`, `text-text-primary` vb.
- Dark: `@custom-variant dark` + `html.dark` class
- Utility override: `!important` sadece `.prose` font inheritance için

Yeni renk eklerken `@theme` bloğuna token ekleyin; dark override'ı `.dark body` altında güncelleyin.

---

## 10. Bileşen dosya haritası

| Bileşen | Dosya | Stil kaynağı |
|---------|-------|--------------|
| Global tema | `app.css` | @theme, body, dark |
| Layout | `App.svelte` | Tailwind utility |
| Toolbar | `Toolbar.svelte` | Tailwind + tooltip inline |
| Editör | `Editor.svelte` | app.css ProseMirror |
| Review kart | `ReviewCard.svelte` | app.css `.review-*` + `.diff-*` |
| Review panel | `ReviewPanel.svelte` | app.css header + toplu aksiyonlar |
| Toast | `Toast.svelte` | Tailwind + slideUp animasyon |
| Yazım popup | `SpellSuggestion.svelte` | app.css `.spell-*` |
| Kurulum | `SetupWizard.svelte` | Tailwind utility |
| Ayarlar | `SettingsDialog.svelte` | Tailwind utility |
| Yorumlar | `CommentPanel.svelte` | Tailwind utility |
| Özet | `SummaryPanel.svelte` | Tailwind utility |

## 11. Klavye kısayolları

| Kısayol | Aksiyon |
|---------|---------|
| `Ctrl+Shift+I` | AI İyileştir (paragraf) |
| `Ctrl+Enter` | Aktif öneriyi onayla |
| `Esc` | Aktif öneriyi reddet |
| `Ctrl+Shift+]` | Sonraki öneri |
| `Ctrl+Shift+[` | Önceki öneri |

## 12. Tasarım kontrol listesi (yeni özellik)

1. Renkler tema token'ından; sabit hex yalnızca tooltip gibi istisnalarda
2. Dark mode'da test edildi
3. Hover-to-reveal gereksiz butonları gizler
4. Diff renkleri pastel; yüksek kontrast kırmızı/yeşil kaçınılır
5. UI metinleri Türkçe
6. Okuma alanı max-w-3xl convention'ına uygun
7. Spacing: border-border ayırıcılar, surface-secondary hover
8. Animasyon 150ms altında; abartılı motion yok
9. Yazım/diff sınıfları app.css'te merkezi tanımlı
