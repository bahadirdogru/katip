<script lang="ts">
  import type { Editor } from '@tiptap/core';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import { commentStore } from '../stores/commentStore.svelte.ts';
  import { settingsStore } from '../stores/settingsStore.svelte.ts';
  import { toastStore } from '../stores/toastStore.svelte.ts';
  import type { ImprovementMode } from '../stores/reviewStore.svelte.ts';
  import type { LLMState } from '../editor/aiReviewService.ts';
  import {
    requestImprovement,
    cancelActiveImprovement,
    collectParagraphs,
    findTextblockAtPos,
  } from '../editor/aiReviewService.ts';
  import { applyDiffForReview } from '../editor/diffDecorations.ts';
  import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js';
  import { documentStore } from '../stores/documentStore.svelte.ts';

  interface Props {
    editor: Editor;
    llmState?: LLMState;
  }

  let { editor, llmState = 'off' }: Props = $props();
  let aiProcessing = $state(false);
  let isScanningAll = $state(false);
  let scanPaused = $state(false);
  let scanProgress = $state({ current: 0, total: 0, found: 0 });
  let selectedMode = $state<ImprovementMode>('fix');
  let showModeMenu = $state(false);

  const modes: { id: ImprovementMode; label: string }[] = [
    { id: 'fix', label: 'Düzelt' },
    { id: 'shorten', label: 'Kısalt' },
    { id: 'flow', label: 'Akıcılaştır' },
    { id: 'formal', label: 'Resmileştir' },
  ];

  let aiReady = $derived(llmState === 'ready');
  let aiDisabledReason = $derived(
    llmState === 'loading' ? 'AI sunucusu yükleniyor...' :
    llmState === 'off' ? 'AI sunucusu kapalı — Ayarlardan başlatın' : ''
  );

  let selectedModeLabel = $derived(modes.find(m => m.id === selectedMode)?.label ?? 'Düzelt');

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
      label: 'Kalın', shortcut: 'Ctrl+B', icon: 'B',
      action: () => (editor.chain().focus() as any).toggleBold().run(),
      isActive: () => editor.isActive('bold'),
    },
    {
      label: 'İtalik', shortcut: 'Ctrl+I', icon: 'I',
      action: () => (editor.chain().focus() as any).toggleItalic().run(),
      isActive: () => editor.isActive('italic'),
    },
    {
      label: 'Altı Çizili', shortcut: 'Ctrl+U', icon: 'U',
      action: () => (editor.chain().focus() as any).toggleUnderline().run(),
      isActive: () => editor.isActive('underline'),
    },
    {
      label: 'Üstü Çizili', shortcut: 'Ctrl+Shift+X', icon: 'S',
      action: () => (editor.chain().focus() as any).toggleStrike().run(),
      isActive: () => editor.isActive('strike'),
    },
    {
      label: 'Vurgulu', shortcut: 'Ctrl+Shift+H', icon: 'A',
      action: () => (editor.chain().focus() as any).toggleHighlight().run(),
      isActive: () => editor.isActive('highlight'),
    },
    { isSeparator: true },
    {
      label: 'Sola Hizala', shortcut: 'Ctrl+Shift+L', icon: '↤',
      action: () => (editor.chain().focus() as any).setTextAlign('left').run(),
      isActive: () => editor.isActive({ textAlign: 'left' }),
    },
    {
      label: 'Ortala', shortcut: 'Ctrl+Shift+E', icon: '↔',
      action: () => (editor.chain().focus() as any).setTextAlign('center').run(),
      isActive: () => editor.isActive({ textAlign: 'center' }),
    },
    {
      label: 'Sağa Hizala', shortcut: 'Ctrl+Shift+R', icon: '↦',
      action: () => (editor.chain().focus() as any).setTextAlign('right').run(),
      isActive: () => editor.isActive({ textAlign: 'right' }),
    },
    { isSeparator: true },
    {
      label: 'Başlık 1', shortcut: 'Ctrl+Alt+1', icon: 'H1',
      action: () => (editor.chain().focus() as any).toggleHeading({ level: 1 }).run(),
      isActive: () => editor.isActive('heading', { level: 1 }),
    },
    {
      label: 'Başlık 2', shortcut: 'Ctrl+Alt+2', icon: 'H2',
      action: () => (editor.chain().focus() as any).toggleHeading({ level: 2 }).run(),
      isActive: () => editor.isActive('heading', { level: 2 }),
    },
    { isSeparator: true },
    {
      label: 'Madde Listesi', shortcut: 'Ctrl+Shift+8', icon: '• List',
      action: () => (editor.chain().focus() as any).toggleBulletList().run(),
      isActive: () => editor.isActive('bulletList'),
    },
    {
      label: 'Yorum Ekle', shortcut: 'Ctrl+Alt+C', icon: '💬',
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
  ]);

  async function handleAIImprove() {
    if (!aiReady || aiProcessing) return;
    aiProcessing = true;
    try {
      await requestImprovement(editor, { scope: 'paragraph', mode: selectedMode, llmState });
    } finally {
      aiProcessing = false;
    }
  }

  async function handleAISelection() {
    if (!aiReady || aiProcessing) return;
    aiProcessing = true;
    try {
      await requestImprovement(editor, { scope: 'selection', mode: selectedMode, llmState });
    } finally {
      aiProcessing = false;
    }
  }

  function handleCancel() {
    cancelActiveImprovement();
    aiProcessing = false;
    isScanningAll = false;
    scanPaused = false;
  }

  async function handleBulkScan() {
    if (!aiReady || aiProcessing || isScanningAll) return;

    const paragraphs = collectParagraphs(editor, 10);
    if (paragraphs.length === 0) {
      toastStore.warning('Taranacak paragraf bulunamadı.');
      return;
    }

    isScanningAll = true;
    scanPaused = false;
    scanProgress = { current: 0, total: paragraphs.length, found: 0 };

    for (const para of paragraphs) {
      if (scanPaused) {
        toastStore.info(`Tarama duraklatıldı (${scanProgress.current}/${scanProgress.total}).`);
        break;
      }

      scanProgress.current++;

      try {
        const result = await ImproveParagraph(para.id, para.text, documentStore.plotSummary, selectedMode);
        if (result?.diffs?.length) {
          const hasChanges = result.diffs.some((d: any) => d.type !== 'equal');
          if (hasChanges && !reviewStore.reviews.some((r) => r.original === para.text && r.status === 'pending')) {
            const mappedDiffs = result.diffs.map((d: any) => ({
              type: d.type as 'equal' | 'insert' | 'delete',
              text: d.text,
            }));

            const reviewId = reviewStore.addReview({
              paragraphId: result.paragraphId,
              paragraphPos: para.pos,
              summary: result.summary,
              original: result.original,
              improved: result.improved,
              diffs: mappedDiffs,
              mode: selectedMode,
              changeType: result.changeType,
            });

            const block = findTextblockAtPos(editor, para.pos);
            if (block) {
              applyDiffForReview(editor, reviewId, result.paragraphId, block.from, mappedDiffs);
            }
            scanProgress.found++;
          }
        }
      } catch (err) {
        console.error('Toplu tarama hatası (paragraf atlandı):', err);
      }
    }

    isScanningAll = false;
    toastStore.success(`${scanProgress.total} paragraftan ${scanProgress.found}'ünde öneri bulundu.`);
  }

  function toggleScanPause() {
    scanPaused = !scanPaused;
  }
</script>

<svelte:window onclick={() => showModeMenu = false} />

<div class="flex items-center gap-1 px-3 py-2 border-b border-border bg-surface shrink-0 overflow-x-auto">
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
        {#if action.icon === 'I'}<span class="italic">{action.icon}</span>
        {:else if action.icon === 'S'}<span class="line-through">{action.icon}</span>
        {:else if action.icon === 'B'}<span class="font-bold">{action.icon}</span>
        {:else if action.icon === 'U'}<span class="underline">{action.icon}</span>
        {:else}{action.icon}{/if}
      </button>
    {/if}
  {/each}

  <div class="w-px h-5 bg-border mx-2"></div>

  <!-- Mode dropdown -->
  <div class="relative">
    <button
      class="px-2 py-1.5 text-xs rounded border border-border bg-surface-secondary text-text-secondary hover:text-text-primary"
      onclick={(e) => { e.stopPropagation(); showModeMenu = !showModeMenu; }}
    >{selectedModeLabel} ▾</button>
    {#if showModeMenu}
      <div class="absolute top-full left-0 mt-1 bg-surface border border-border rounded shadow-lg z-50 min-w-[120px]">
        {#each modes as mode}
          <button
            class="block w-full text-left px-3 py-1.5 text-xs hover:bg-surface-secondary {selectedMode === mode.id ? 'text-primary font-medium' : 'text-text-secondary'}"
            onclick={(e) => { e.stopPropagation(); selectedMode = mode.id; showModeMenu = false; }}
          >{mode.label}</button>
        {/each}
      </div>
    {/if}
  </div>

  <button
    class="flex items-center gap-2 px-4 py-1.5 text-sm font-medium rounded-md transition-all disabled:opacity-40
      {aiProcessing ? 'text-amber-700 bg-amber-100 dark:text-amber-300 dark:bg-amber-900/40' : 'text-white bg-primary hover:bg-primary-dark shadow-sm hover:shadow'}"
    onclick={handleAIImprove}
    disabled={!aiReady || aiProcessing}
    title={aiDisabledReason || `İmlecin bulunduğu paragrafı ${selectedModeLabel} modunda iyileştir (Ctrl+Shift+I)`}
  >
    {#if aiProcessing}<span class="animate-spin">⏳</span><span>Analiz ediliyor...</span>
    {:else}<span>✨</span><span>AI İyileştir</span>{/if}
  </button>

  <button
    class="px-3 py-1.5 text-xs rounded border border-border text-text-secondary hover:bg-surface-secondary disabled:opacity-40"
    onclick={handleAISelection}
    disabled={!aiReady || aiProcessing}
    title="Seçili metni AI ile iyileştir"
  >Seçimi İyileştir</button>

  {#if aiProcessing}
    <button class="px-2 py-1.5 text-xs rounded text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30" onclick={handleCancel}>İptal</button>
  {/if}

  <button
    class="flex items-center gap-2 px-4 py-1.5 text-sm font-medium rounded-md transition-all disabled:opacity-40
      {isScanningAll ? 'text-blue-700 bg-blue-100 dark:text-blue-300 dark:bg-blue-900/40' : 'text-text-secondary bg-surface-secondary hover:bg-surface border border-border shadow-sm'}"
    onclick={handleBulkScan}
    disabled={!aiReady || aiProcessing || isScanningAll}
    title={aiDisabledReason || 'Belgedeki tüm paragrafları sırayla tarayın'}
  >
    {#if isScanningAll}
      <span class="animate-spin">⏳</span>
      <span>{scanProgress.current} / {scanProgress.total}</span>
    {:else}
      <span>🔎</span><span>Tümünü Tara</span>
    {/if}
  </button>

  {#if isScanningAll}
    <button class="px-2 py-1.5 text-xs rounded border border-border text-text-secondary hover:bg-surface-secondary" onclick={toggleScanPause}>
      {scanPaused ? 'Devam' : 'Durdur'}
    </button>
  {/if}
</div>
