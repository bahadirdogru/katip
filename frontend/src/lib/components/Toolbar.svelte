<script lang="ts">
  import type { Editor } from '@tiptap/core';
  import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import { buildDecorations, diffPluginKey } from '../editor/diffDecorations.ts';

  interface Props {
    editor: Editor;
  }

  let { editor }: Props = $props();
  let aiProcessing = $state(false);

  type ToolbarAction = {
    label: string;
    icon: string;
    action: () => void;
    isActive: () => boolean;
  };

  let actions: ToolbarAction[] = $derived([
    {
      label: 'Kalın',
      icon: 'B',
      action: () => editor.chain().focus().toggleBold().run(),
      isActive: () => editor.isActive('bold'),
    },
    {
      label: 'İtalik',
      icon: 'I',
      action: () => editor.chain().focus().toggleItalic().run(),
      isActive: () => editor.isActive('italic'),
    },
    {
      label: 'Üstü Çizili',
      icon: 'S',
      action: () => editor.chain().focus().toggleStrike().run(),
      isActive: () => editor.isActive('strike'),
    },
    {
      label: 'Başlık 1',
      icon: 'H1',
      action: () => editor.chain().focus().toggleHeading({ level: 1 }).run(),
      isActive: () => editor.isActive('heading', { level: 1 }),
    },
    {
      label: 'Başlık 2',
      icon: 'H2',
      action: () => editor.chain().focus().toggleHeading({ level: 2 }).run(),
      isActive: () => editor.isActive('heading', { level: 2 }),
    },
    {
      label: 'Başlık 3',
      icon: 'H3',
      action: () => editor.chain().focus().toggleHeading({ level: 3 }).run(),
      isActive: () => editor.isActive('heading', { level: 3 }),
    },
    {
      label: 'Madde Listesi',
      icon: '•',
      action: () => editor.chain().focus().toggleBulletList().run(),
      isActive: () => editor.isActive('bulletList'),
    },
    {
      label: 'Numaralı Liste',
      icon: '1.',
      action: () => editor.chain().focus().toggleOrderedList().run(),
      isActive: () => editor.isActive('orderedList'),
    },
    {
      label: 'Alıntı',
      icon: '"',
      action: () => editor.chain().focus().toggleBlockquote().run(),
      isActive: () => editor.isActive('blockquote'),
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
      const result = await ImproveParagraph(`p-${para.pos}`, para.text);
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

<div class="flex items-center gap-0.5 px-3 py-1 border-b border-border bg-surface shrink-0 overflow-x-auto">
  {#each actions as action}
    <button
      class="px-1.5 py-0.5 text-xs rounded transition-colors min-w-[24px]
        {action.isActive() ? 'text-primary bg-primary/10' : 'text-text-secondary hover:bg-surface-secondary'}"
      onclick={action.action}
      title={action.label}
    >
      {#if action.icon === 'I'}
        <span class="italic">{action.icon}</span>
      {:else if action.icon === 'S'}
        <span class="line-through">{action.icon}</span>
      {:else if action.icon === 'B'}
        <span class="font-bold">{action.icon}</span>
      {:else}
        {action.icon}
      {/if}
    </button>
  {/each}

  <div class="w-px h-4 bg-border mx-1"></div>

  <button
    class="px-1.5 py-0.5 text-xs rounded text-text-secondary hover:bg-surface-secondary disabled:opacity-30"
    onclick={() => editor.chain().focus().undo().run()}
    disabled={!editor.can().undo()}
    title="Geri Al"
  >
    ↩
  </button>
  <button
    class="px-1.5 py-0.5 text-xs rounded text-text-secondary hover:bg-surface-secondary disabled:opacity-30"
    onclick={() => editor.chain().focus().redo().run()}
    disabled={!editor.can().redo()}
    title="Yinele"
  >
    ↪
  </button>

  <div class="w-px h-4 bg-border mx-1"></div>

  <button
    class="flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-md transition-all disabled:opacity-40
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
</div>
