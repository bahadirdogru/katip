<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Editor from './lib/components/Editor.svelte';
  import Toolbar from './lib/components/Toolbar.svelte';
  import ReviewPanel from './lib/components/ReviewPanel.svelte';
  import SettingsDialog from './lib/components/SettingsDialog.svelte';
  import SetupWizard from './lib/components/SetupWizard.svelte';
  import { CheckSetupStatus, GetLLMStatus } from '../bindings/katip/internal/service/katipservice.js';
  import { reviewStore } from './lib/stores/reviewStore.svelte.ts';
  import { diffPluginKey } from './lib/editor/diffDecorations.ts';
  import type { Editor as TipTapEditor } from '@tiptap/core';

  let editor: TipTapEditor | null = $state(null);
  let showReviewPanel = $state(true);
  let showSettings = $state(false);

  let setupInfo = $state<any>(null);
  let loading = $state(true);

  type LLMState = 'off' | 'loading' | 'ready';
  let llmState = $state<LLMState>('off');
  let statusPollTimer: ReturnType<typeof setInterval> | null = null;

  let dark = $state(false);

  function applyTheme(isDark: boolean) {
    document.documentElement.classList.toggle('dark', isDark);
    localStorage.setItem('katip-theme', isDark ? 'dark' : 'light');
  }

  function toggleTheme() {
    dark = !dark;
    applyTheme(dark);
  }

  async function pollLLMStatus() {
    try {
      const status = await GetLLMStatus() as any;
      if (status?.healthy) {
        llmState = 'ready';
      } else if (status?.running) {
        llmState = 'loading';
      } else {
        llmState = 'off';
      }
    } catch {
      llmState = 'off';
    }
  }

  onMount(async () => {
    const saved = localStorage.getItem('katip-theme');
    dark = saved === 'dark';
    applyTheme(dark);

    try {
      setupInfo = await CheckSetupStatus();
    } catch (e) {
      console.error('Setup durumu kontrol edilemedi:', e);
      setupInfo = { status: 'ready' };
    }
    loading = false;

    pollLLMStatus();
    statusPollTimer = setInterval(pollLLMStatus, 5000);
  });

  onDestroy(() => {
    if (statusPollTimer) clearInterval(statusPollTimer);
  });

  async function handleSetupComplete() {
    setupInfo = await CheckSetupStatus();
  }

  function handleEditorReady(e: TipTapEditor) {
    editor = e;
  }

  function clearDecorations() {
    if (!editor) return;
    const tr = editor.view.state.tr.setMeta(diffPluginKey, { clear: true });
    editor.view.dispatch(tr);
  }

  function handleAccept(reviewId: string) {
    const review = reviewStore.reviews.find(r => r.id === reviewId);
    if (review && editor) {
      clearDecorations();

      const { state, dispatch } = editor.view;
      let found = false;

      state.doc.descendants((node, pos) => {
        if (found) return false;
        if (node.isTextblock && node.textContent === review.original) {
          const tr = state.tr.replaceWith(
            pos + 1,
            pos + node.nodeSize - 1,
            state.schema.text(review.improved)
          );
          dispatch(tr);
          found = true;
          return false;
        }
      });

      reviewStore.acceptReview(reviewId);
    }
  }

  function handleReject(reviewId: string) {
    clearDecorations();
    reviewStore.rejectReview(reviewId);
  }

  let hasReviews = $derived(reviewStore.reviews.length > 0);
</script>

{#if loading}
  <div class="flex items-center justify-center h-screen w-screen bg-surface">
    <div class="text-sm text-text-secondary">Yükleniyor...</div>
  </div>
{:else if setupInfo?.status !== 'ready'}
  <SetupWizard {setupInfo} onComplete={handleSetupComplete} />
{:else}
  <div class="flex flex-col h-screen w-screen">
    <header class="flex items-center justify-between px-4 py-1.5 border-b border-border bg-surface shrink-0"
      style="--wails-draggable: drag;">
      <div class="flex items-center gap-3">
        <h1 class="text-sm font-semibold text-text-primary tracking-tight">Katip</h1>
        <button
          class="flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] transition-colors hover:bg-surface-secondary"
          onclick={() => showSettings = true}
          title={llmState === 'ready' ? 'AI Sunucusu hazır' : llmState === 'loading' ? 'AI Sunucusu yükleniyor...' : 'AI Sunucusu kapalı — Ayarlardan başlatın'}
        >
          <span class="relative flex h-2.5 w-2.5">
            {#if llmState === 'loading'}
              <span class="absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75 animate-ping"></span>
            {/if}
            <span class="relative inline-flex h-2.5 w-2.5 rounded-full {llmState === 'ready' ? 'bg-green-500' : llmState === 'loading' ? 'bg-amber-400' : 'bg-red-400'}"></span>
          </span>
          <span class="text-text-secondary">
            {llmState === 'ready' ? 'AI Hazır' : llmState === 'loading' ? 'Yükleniyor' : 'AI Kapalı'}
          </span>
        </button>
      </div>
      <div class="flex items-center gap-1">
        <button
          class="text-xs px-2.5 py-1 rounded transition-colors
            {showReviewPanel ? 'text-primary bg-primary/10' : 'text-text-secondary hover:bg-surface-secondary'}"
          onclick={() => showReviewPanel = !showReviewPanel}
        >
          Düzeltmeler{hasReviews ? ` (${reviewStore.reviews.length})` : ''}
        </button>
        <button
          class="text-xs px-2 py-1 rounded text-text-secondary hover:bg-surface-secondary transition-colors"
          onclick={toggleTheme}
          title={dark ? 'Gündüz modu' : 'Gece modu'}
        >
          {dark ? '☀️' : '🌙'}
        </button>
        <button
          class="text-xs px-2 py-1 rounded text-text-secondary hover:bg-surface-secondary transition-colors"
          onclick={() => showSettings = true}
          title="Ayarlar"
        >
          ⚙
        </button>
      </div>
    </header>

    {#if editor}
      <Toolbar {editor} />
    {/if}

    <div class="flex flex-1 overflow-hidden">
      <main class="flex-1 overflow-y-auto bg-surface">
        <div class="max-w-3xl mx-auto px-20 py-8">
          <Editor onReady={handleEditorReady} />
        </div>
      </main>

      {#if showReviewPanel}
        <aside class="w-72 border-l border-border bg-surface overflow-y-auto shrink-0">
          <ReviewPanel {editor} onAccept={handleAccept} onReject={handleReject} />
        </aside>
      {/if}
    </div>
  </div>

  <SettingsDialog open={showSettings} onClose={() => showSettings = false} />
{/if}
