<script lang="ts">
  import type { Editor } from '@tiptap/core';
  import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import { commentStore } from '../stores/commentStore.svelte.ts';
  import { documentStore } from '../stores/documentStore.svelte.ts';
  import { settingsStore } from '../stores/settingsStore.svelte.ts';
  import { buildDecorations, diffPluginKey } from '../editor/diffDecorations.ts';

  interface Props {
    editor: Editor;
  }

  let { editor }: Props = $props();
  let aiProcessing = $state(false);
  let isScanningAll = $state(false);
  let scanProgress = $state({ current: 0, total: 0 });

  type ToolbarItem = {
    isSeparator?: boolean;
    label?: string;
    shortcut?: string;
    icon?: string;
    action?: () => void;
    isActive?: () => boolean;
  };

  let actions: ToolbarItem[] = $derived([
    {
      label: 'Kalın',
      shortcut: 'Ctrl+B',
      icon: 'B',
      action: () => (editor.chain().focus() as any).toggleBold().run(),
      isActive: () => editor.isActive('bold'),
    },
    {
      label: 'İtalik',
      shortcut: 'Ctrl+I',
      icon: 'I',
      action: () => (editor.chain().focus() as any).toggleItalic().run(),
      isActive: () => editor.isActive('italic'),
    },
    {
      label: 'Altı Çizili',
      shortcut: 'Ctrl+U',
      icon: 'U',
      action: () => (editor.chain().focus() as any).toggleUnderline().run(),
      isActive: () => editor.isActive('underline'),
    },
    {
      label: 'Üstü Çizili',
      shortcut: 'Ctrl+Shift+X',
      icon: 'S',
      action: () => (editor.chain().focus() as any).toggleStrike().run(),
      isActive: () => editor.isActive('strike'),
    },
    {
      label: 'Vurgulu',
      shortcut: 'Ctrl+Shift+H',
      icon: 'A',
      action: () => (editor.chain().focus() as any).toggleHighlight().run(),
      isActive: () => editor.isActive('highlight'),
    },
    { isSeparator: true },
    {
      label: 'Sola Hizala',
      shortcut: 'Ctrl+Shift+L',
      icon: '↤',
      action: () => (editor.chain().focus() as any).setTextAlign('left').run(),
      isActive: () => editor.isActive({ textAlign: 'left' }),
    },
    {
      label: 'Ortala',
      shortcut: 'Ctrl+Shift+E',
      icon: '↔',
      action: () => (editor.chain().focus() as any).setTextAlign('center').run(),
      isActive: () => editor.isActive({ textAlign: 'center' }),
    },
    {
      label: 'Sağa Hizala',
      shortcut: 'Ctrl+Shift+R',
      icon: '↦',
      action: () => (editor.chain().focus() as any).setTextAlign('right').run(),
      isActive: () => editor.isActive({ textAlign: 'right' }),
    },
    {
      label: 'İki Yana Yasla',
      shortcut: 'Ctrl+Shift+J',
      icon: '⇿',
      action: () => (editor.chain().focus() as any).setTextAlign('justify').run(),
      isActive: () => editor.isActive({ textAlign: 'justify' }),
    },
    { isSeparator: true },
    {
      label: 'Başlık 1',
      shortcut: 'Ctrl+Alt+1',
      icon: 'H1',
      action: () => (editor.chain().focus() as any).toggleHeading({ level: 1 }).run(),
      isActive: () => editor.isActive('heading', { level: 1 }),
    },
    {
      label: 'Başlık 2',
      shortcut: 'Ctrl+Alt+2',
      icon: 'H2',
      action: () => (editor.chain().focus() as any).toggleHeading({ level: 2 }).run(),
      isActive: () => editor.isActive('heading', { level: 2 }),
    },
    {
      label: 'Başlık 3',
      shortcut: 'Ctrl+Alt+3',
      icon: 'H3',
      action: () => (editor.chain().focus() as any).toggleHeading({ level: 3 }).run(),
      isActive: () => editor.isActive('heading', { level: 3 }),
    },
    { isSeparator: true },
    {
      label: 'Madde Listesi',
      shortcut: 'Ctrl+Shift+8',
      icon: '• List',
      action: () => (editor.chain().focus() as any).toggleBulletList().run(),
      isActive: () => editor.isActive('bulletList'),
    },
    {
      label: 'Numaralı Liste',
      shortcut: 'Ctrl+Shift+7',
      icon: '1. List',
      action: () => (editor.chain().focus() as any).toggleOrderedList().run(),
      isActive: () => editor.isActive('orderedList'),
    },
    {
      label: 'Görev Listesi',
      shortcut: 'Ctrl+Shift+9',
      icon: '☑ Task',
      action: () => (editor.chain().focus() as any).toggleTaskList().run(),
      isActive: () => editor.isActive('taskList'),
    },
    {
      label: 'Alıntı',
      shortcut: 'Ctrl+Shift+B',
      icon: '""',
      action: () => (editor.chain().focus() as any).toggleBlockquote().run(),
      isActive: () => editor.isActive('blockquote'),
    },
    { isSeparator: true },
    {
      label: 'Yorum Ekle',
      shortcut: 'Ctrl+Alt+C',
      icon: '💬',
      action: () => {
        const { state } = editor.view;
        const { from, to, empty } = state.selection;
        if (empty) return;
        
        const text = state.doc.textBetween(from, to, ' ');
        const id = crypto.randomUUID();
        commentStore.addThread(id, text, 'Yeni yorumunuz...', 'Yazar');
        (editor.chain().focus() as any).setComment(id).run();
      },
      isActive: () => editor.isActive('comment'),
    },
    { isSeparator: true },
    {
      label: 'Görsel Ekle',
      shortcut: '',
      icon: '🖼️',
      action: () => {
        const url = window.prompt('Görsel URL:');
        if (url) {
          (editor.chain().focus() as any).setImage({ src: url }).run();
        }
      },
      isActive: () => editor.isActive('image'),
    },
    {
      label: 'Tablo Ekle',
      shortcut: '',
      icon: '▦',
      action: () => (editor.chain().focus() as any).insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(),
      isActive: () => editor.isActive('table'),
    },
    {
      label: 'Sütun Ekle',
      shortcut: '',
      icon: '⊞',
      action: () => (editor.chain().focus() as any).addColumnAfter().run(),
      isActive: () => editor.isActive('table'),
    },
    {
      label: 'Satır Ekle',
      shortcut: '',
      icon: '➕',
      action: () => (editor.chain().focus() as any).addRowAfter().run(),
      isActive: () => editor.isActive('table'),
    },
    {
      label: 'Metin Rengi',
      shortcut: '',
      icon: '🖍️',
      action: () => {
        const color = window.prompt('Renk kodu (örn: #ef4444 veya red):', '#000000');
        if (color) {
          (editor.chain().focus() as any).setColor(color).run();
        }
      },
      isActive: () => editor.isActive('textStyle'),
    },
  ]);

  function getCurrentParagraphText(): { text: string; pos: number } | null {
    const { state } = editor.view;
    const resolved = state.selection.$from;
    const node = resolved.parent;
    if (node.isTextblock && node.textContent.trim().length >= 3) {
      const pos = resolved.before(resolved.depth);
      return { text: node.textContent.trim(), pos };
    }
    return null;
  }

  async function handleAIImprove() {
    const para = getCurrentParagraphText();
    if (!para || aiProcessing) return;

    aiProcessing = true;
    try {
      const result = await ImproveParagraph(`p-${para.pos}`, para.text, documentStore.plotSummary);
      if (result?.diffs?.length) {
        const hasChanges = result.diffs.some((d: any) => d.type !== 'equal');
        if (hasChanges) {
          const mappedDiffs = result.diffs.map((d: any) => ({
            type: d.type as 'equal' | 'insert' | 'delete',
            text: d.text,
          }));

          reviewStore.addReview({
            paragraphId: result.paragraphId,
            summary: result.summary,
            original: result.original,
            improved: result.improved,
            diffs: mappedDiffs,
          });

          applyDiffDecorations(para.text, mappedDiffs);
        }
      }
    } catch (err) {
      console.error('AI iyileştirme hatası:', err);
    } finally {
      aiProcessing = false;
    }
  }

  async function handleBulkScan() {
    if (aiProcessing || isScanningAll) return;
    
    const { state } = editor.view;
    const paragraphs: { id: string; text: string; pos: number }[] = [];
    
    state.doc.descendants((node, pos) => {
      if (node.isTextblock) {
         const text = node.textContent.trim();
         if (text.length >= 10) {
           paragraphs.push({ id: `p-${pos}`, text, pos });
         }
      }
      return false; 
    });

    if (paragraphs.length === 0) return;

    isScanningAll = true;
    scanProgress = { current: 0, total: paragraphs.length };

    for (const para of paragraphs) {
      scanProgress.current++;
      
      try {
        const result = await ImproveParagraph(para.id, para.text, documentStore.plotSummary);
        if (result?.diffs?.length) {
          const hasChanges = result.diffs.some((d: any) => d.type !== 'equal');
          if (hasChanges) {
            const mappedDiffs = result.diffs.map((d: any) => ({
              type: d.type as 'equal' | 'insert' | 'delete',
              text: d.text,
            }));

            if (!reviewStore.reviews.some((r) => r.original === para.text)) {
                reviewStore.addReview({
                  paragraphId: result.paragraphId,
                  summary: result.summary,
                  original: result.original,
                  improved: result.improved,
                  diffs: mappedDiffs,
                });
                applyDiffDecorations(para.text, mappedDiffs);
            }
          }
        }
      } catch (err) {
        console.error('Toplu tarama hatası (paragraf atlandı):', err);
      }
    }
    
    isScanningAll = false;
  }

  function applyDiffDecorations(originalText: string, diffs: Array<{type: 'equal'|'insert'|'delete', text: string}>) {
    const { state } = editor.view;
    let paragraphFrom: number | null = null;

    state.doc.descendants((node, pos) => {
      if (paragraphFrom !== null) return false;
      if (node.isTextblock && node.textContent === originalText) {
        paragraphFrom = pos + 1;
        return false;
      }
    });

    if (paragraphFrom === null) return;

    const decorations = buildDecorations(state.doc, paragraphFrom, diffs);
    const tr = state.tr.setMeta(diffPluginKey, { decorations });
    editor.view.dispatch(tr);
  }
</script>

<div class="flex items-center gap-1 px-3 py-2 border-b border-border bg-surface shrink-0 overflow-x-auto">
  <!-- Zoom Controls -->
  <div class="flex items-center bg-surface-secondary border border-border rounded px-1 py-1 mr-2 gap-1">
    <button class="text-xs font-medium px-2 py-0.5 rounded text-text-secondary hover:text-primary hover:bg-surface border border-transparent hover:border-border transition-colors focus:outline-none"
      onclick={() => settingsStore.setFontSize(settingsStore.fontSize - 1)} title="Metni Küçült">A-</button>
    <span class="text-xs w-6 text-center text-text-primary font-mono">{settingsStore.fontSize}</span>
    <button class="text-xs font-medium px-2 py-0.5 rounded text-text-secondary hover:text-primary hover:bg-surface border border-transparent hover:border-border transition-colors focus:outline-none"
      onclick={() => settingsStore.setFontSize(settingsStore.fontSize + 1)} title="Metni Büyüt">A+</button>
  </div>
  
  <div class="w-px h-4 bg-border mx-1"></div>

  {#each actions as action}
    {#if action.isSeparator}
      <div class="w-px h-4 bg-border mx-1"></div>
    {:else}
      <button
        class="group relative px-2.5 py-1.5 text-sm rounded transition-colors min-w-[32px] flex items-center justify-center shadow-sm
          {action.isActive && action.isActive() ? 'text-primary bg-primary/15' : 'text-text-secondary hover:bg-surface-secondary hover:text-text-primary'}"
        onclick={action.action}
      >
        {#if action.icon === 'I'}
          <span class="italic">{action.icon}</span>
        {:else if action.icon === 'S'}
          <span class="line-through">{action.icon}</span>
        {:else if action.icon === 'B'}
          <span class="font-bold">{action.icon}</span>
        {:else if action.icon === 'U'}
          <span class="underline">{action.icon}</span>
        {:else if action.icon === 'A'}
          <span class="bg-yellow-200 dark:bg-yellow-800/50 rounded px-0.5">{action.icon}</span>
        {:else}
          {action.icon}
        {/if}

        <!-- Tooltip -->
        <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2.5 py-1 bg-gray-800 text-white text-xs whitespace-nowrap rounded opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all pointer-events-none z-50">
          <div class="font-medium">{action.label}</div>
          <div class="text-gray-400 font-mono text-[9px] mt-0.5 opacity-90">{action.shortcut}</div>
          <div class="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-l-transparent border-r-transparent border-t-gray-800"></div>
        </div>
      </button>
    {/if}
  {/each}

  <div class="w-px h-4 bg-border mx-1"></div>

  <button
    class="px-2.5 py-1.5 text-sm rounded text-text-secondary hover:bg-surface-secondary hover:text-text-primary disabled:opacity-30"
    onclick={() => editor.chain().focus().undo().run()}
    disabled={!editor.can().undo()}
    title="Geri Al"
  >
    ↩
  </button>
  <button
    class="px-2.5 py-1.5 text-sm rounded text-text-secondary hover:bg-surface-secondary hover:text-text-primary disabled:opacity-30"
    onclick={() => editor.chain().focus().redo().run()}
    disabled={!editor.can().redo()}
    title="Yinele"
  >
    ↪
  </button>

  <div class="w-px h-5 bg-border mx-2"></div>

  <button
    class="flex items-center gap-2 px-4 py-1.5 text-sm font-medium rounded-md transition-all disabled:opacity-40
      {aiProcessing
        ? 'text-amber-700 bg-amber-100 dark:text-amber-300 dark:bg-amber-900/40'
        : 'text-white bg-primary hover:bg-primary-dark shadow-sm hover:shadow'}"
    onclick={handleAIImprove}
    disabled={aiProcessing}
    title="İmlecin bulunduğu paragrafı AI ile iyileştir"
  >
    {#if aiProcessing}
      <span class="animate-spin">⏳</span>
      <span>Analiz ediliyor...</span>
    {:else}
      <span>✨</span>
      <span>AI İyileştir</span>
    {/if}
  </button>
  
  <button
    class="flex items-center gap-2 px-4 py-1.5 text-sm font-medium rounded-md transition-all disabled:opacity-40
      {isScanningAll
        ? 'text-blue-700 bg-blue-100 dark:text-blue-300 dark:bg-blue-900/40'
        : 'text-text-secondary bg-surface-secondary hover:bg-surface border border-border shadow-sm'}"
    onclick={handleBulkScan}
    disabled={aiProcessing || isScanningAll}
    title="Belgedeki tüm yazıları sırayla tarayarak yapay zekaya denetlet."
  >
    {#if isScanningAll}
      <span class="animate-spin">⏳</span>
      <span>{scanProgress.current} / {scanProgress.total}</span>
    {:else}
      <span>🔎</span>
      <span>Tümünü Tara</span>
    {/if}
  </button>
</div>
