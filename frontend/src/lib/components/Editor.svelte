<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  import Placeholder from '@tiptap/extension-placeholder';
  import { ImproveParagraph } from '../../../bindings/katip/internal/service/katipservice.js';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import { createDiffPlugin, diffPluginKey, buildDecorations } from '../editor/diffDecorations.ts';
  import { createSpellcheckPlugin, spellPluginKey } from '../editor/spellcheckPlugin.ts';
  import { initSpellChecker, isReady as isSpellReady } from '../editor/spellChecker.ts';
  import SpellSuggestion from './SpellSuggestion.svelte';

  interface Props {
    onReady?: (editor: Editor) => void;
  }

  let { onReady }: Props = $props();

  let element: HTMLDivElement | undefined = $state();
  let editor: Editor | undefined = $state();
  let hoverButtonEl: HTMLDivElement | undefined = $state();
  let hoverParagraph: { node: Element; pos: number; text: string } | null = $state(null);
  let isProcessing = $state(false);

  let spellPopup: { word: string; x: number; y: number; from: number; to: number } | null = $state(null);

  function showAIButton(event: MouseEvent) {
    if (!editor || !hoverButtonEl || isProcessing) return;

    const target = event.target as HTMLElement;
    const paragraph = target.closest('.ProseMirror p, .ProseMirror h1, .ProseMirror h2, .ProseMirror h3');
    if (!paragraph) {
      hoverParagraph = null;
      return;
    }

    const text = paragraph.textContent?.trim() || '';
    if (text.length < 5) {
      hoverParagraph = null;
      return;
    }

    const view = editor.view;
    const pos = view.posAtDOM(paragraph, 0);
    hoverParagraph = { node: paragraph, pos, text };
  }

  async function handleImprove() {
    if (!hoverParagraph || !editor || isProcessing) return;

    isProcessing = true;
    const paragraphId = `p-${hoverParagraph.pos}`;
    const originalText = hoverParagraph.text;

    try {
      const result = await ImproveParagraph(paragraphId, originalText);
      if (result && result.diffs && result.diffs.length > 0) {
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

          applyDiffDecorations(originalText, mappedDiffs);
        }
      }
    } catch (err) {
      console.error('AI iyileştirme hatası:', err);
    } finally {
      isProcessing = false;
    }
  }

  function applyDiffDecorations(originalText: string, diffs: Array<{type: 'equal'|'insert'|'delete', text: string}>) {
    if (!editor) return;
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

  function handleContextMenu(event: MouseEvent) {
    if (!editor) return;

    const target = event.target as HTMLElement;
    if (!target.closest('.spell-error')) {
      spellPopup = null;
      return;
    }

    event.preventDefault();

    const coords = { left: event.clientX, top: event.clientY };
    const pos = editor.view.posAtCoords(coords);
    if (!pos) return;

    const resolvedPos = editor.view.state.doc.resolve(pos.pos);
    const textNode = resolvedPos.parent;
    if (!textNode.isTextblock) return;

    const blockStart = resolvedPos.start();
    const text = textNode.textContent;
    const offsetInBlock = pos.pos - blockStart;

    const wordRe = /[a-zA-ZçÇğĞıİöÖşŞüÜâÂîÎûÛ]+/g;
    let match: RegExpExecArray | null;
    while ((match = wordRe.exec(text)) !== null) {
      if (match.index <= offsetInBlock && offsetInBlock <= match.index + match[0].length) {
        const editorRect = element?.getBoundingClientRect();
        spellPopup = {
          word: match[0],
          x: event.clientX - (editorRect?.left ?? 0),
          y: event.clientY - (editorRect?.top ?? 0) + (element?.scrollTop ?? 0),
          from: blockStart + match.index,
          to: blockStart + match.index + match[0].length,
        };
        return;
      }
    }
  }

  function handleSpellReplace(newWord: string) {
    if (!editor || !spellPopup) return;
    editor.chain()
      .focus()
      .command(({ tr }) => {
        tr.replaceWith(spellPopup!.from, spellPopup!.to,
          editor!.view.state.schema.text(newWord));
        return true;
      })
      .run();
    spellPopup = null;
  }

  function handleSpellClose() {
    spellPopup = null;
    if (editor) {
      editor.view.dispatch(
        editor.view.state.tr.setMeta(spellPluginKey, { forceUpdate: true })
      );
    }
  }

  onMount(() => {
    if (!element) return;

    const diffPlugin = createDiffPlugin();
    const spellPlugin = createSpellcheckPlugin();

    editor = new Editor({
      element,
      extensions: [
        StarterKit,
        Placeholder.configure({
          placeholder: 'Yazmaya başlayın...',
        }),
      ],
      content: `
        <h2>Katip'e Hoş Geldiniz</h2>
        <p>Bu bir profesyonel Türkçe metin düzenleyicisidir. Paragraflarınızı AI ile iyileştirebilirsiniz.</p>
        <p>Herhangi bir paragrafın üzerine gelin ve sağda beliren "AI İyileştir" butonuna tıklayın. AI, paragrafınızı analiz edecek ve düzeltme önerilerini sağ panelde gösterecektir.</p>
        <p>Önerileri onaylayabilir veya reddedebilirsiniz. Onaylanan değişiklikler metne uygulanır, reddedilen değişiklikler geri alınır.</p>
      `,
      editorProps: {
        attributes: {
          class: 'prose prose-slate max-w-none focus:outline-none',
          spellcheck: 'false',
        },
      },
      onTransaction: ({ editor: updatedEditor }) => {
        editor = updatedEditor;
      },
    });

    editor.registerPlugin(diffPlugin);
    editor.registerPlugin(spellPlugin);

    initSpellChecker().then(() => {
      if (editor) {
        editor.view.dispatch(
          editor.view.state.tr.setMeta(spellPluginKey, { forceUpdate: true })
        );
      }
    }).catch(() => {});

    if (onReady && editor) {
      onReady(editor);
    }
  });

  onDestroy(() => {
    editor?.destroy();
  });
</script>

<div class="relative h-full">
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="editor-container h-full"
    bind:this={element}
    onmousemove={showAIButton}
    onmouseleave={() => hoverParagraph = null}
    oncontextmenu={handleContextMenu}
  ></div>

  {#if hoverParagraph && !isProcessing}
    {@const editorRect = element?.getBoundingClientRect()}
    {@const nodeRect = hoverParagraph.node.getBoundingClientRect()}
    {#if editorRect && nodeRect}
      <div
        bind:this={hoverButtonEl}
        class="absolute z-20"
        style="top: {nodeRect.top - editorRect.top + (element?.scrollTop ?? 0) + 2}px; left: 2px; transform: translateX(-100%);"
      >
        <button
          class="px-1.5 py-0.5 text-[11px] rounded text-text-secondary hover:text-primary hover:bg-primary/10 transition-colors"
          onclick={handleImprove}
          title="AI ile iyileştir"
        >
          İyileştir
        </button>
      </div>
    {/if}
  {/if}

  {#if isProcessing}
    <div class="absolute top-2 left-1/2 -translate-x-1/2 z-20">
      <div class="flex items-center gap-1.5 px-3 py-1 text-xs rounded bg-surface-secondary text-text-secondary border border-border">
        <span class="animate-spin">⏳</span>
        <span>Analiz ediliyor...</span>
      </div>
    </div>
  {/if}

  {#if spellPopup}
    <SpellSuggestion
      word={spellPopup.word}
      x={spellPopup.x}
      y={spellPopup.y}
      onReplace={handleSpellReplace}
      onClose={handleSpellClose}
    />
  {/if}
</div>
