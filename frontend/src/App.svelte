<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Editor from './lib/components/Editor.svelte';
  import Toolbar from './lib/components/Toolbar.svelte';
  import ReviewPanel from './lib/components/ReviewPanel.svelte';
  import SettingsDialog from './lib/components/SettingsDialog.svelte';
  import SetupWizard from './lib/components/SetupWizard.svelte';
  import Toast from './lib/components/Toast.svelte';
  import { CheckSetupStatus, GetLLMStatus, GetConfig, StartLLMServer, GetSystemMonitor } from '../bindings/katip/internal/service/katipservice.js';
  import { reviewStore } from './lib/stores/reviewStore.svelte.ts';
  import { settingsStore } from './lib/stores/settingsStore.svelte.ts';
  import { commentStore } from './lib/stores/commentStore.svelte.ts';
  import CommentPanel from './lib/components/CommentPanel.svelte';
  import SummaryPanel from './lib/components/SummaryPanel.svelte';
  import type { Editor as TipTapEditor } from '@tiptap/core';
  import {
    acceptReview,
    rejectReview,
    scrollToReview,
    updateGranularDecorations,
    requestImprovement,
  } from './lib/editor/aiReviewService.ts';

  let editor: TipTapEditor | null = $state(null);

  type RightPanelTab = 'reviews' | 'comments' | 'summary' | 'closed';
  let activePanel = $state<RightPanelTab>('reviews');

  let showSettings = $state(false);

  let setupInfo = $state<any>(null);
  let loading = $state(true);
  let wizardSkipped = $state(false);

  type LLMState = 'off' | 'loading' | 'ready';
  let llmState = $state<LLMState>('off');
  let statusPollTimer: ReturnType<typeof setInterval> | null = null;

  let dark = $state(false);

  let systemMonitor = $state<{
    totalRAMBytes: number;
    availableRAMBytes: number;
    usedRAMBytes: number;
    cpuCores: number;
    llmRunning: boolean;
    llmHealthy: boolean;
  } | null>(null);
  let monitorTimer: ReturnType<typeof setInterval> | null = null;
  let autoStartAttempted = $state(false);

  function applyTheme(isDark: boolean) {
    document.documentElement.classList.toggle('dark', isDark);
    localStorage.setItem('katip-theme', isDark ? 'dark' : 'light');
  }

  function toggleTheme() {
    dark = !dark;
    applyTheme(dark);
  }

  async function pollSystemMonitor() {
    try {
      systemMonitor = await GetSystemMonitor() as any;
    } catch {
      systemMonitor = null;
    }
  }

  async function tryAutoStartLLM() {
    if (autoStartAttempted) return;
    autoStartAttempted = true;
    try {
      const setup = await CheckSetupStatus() as any;
      if (setup?.status !== 'ready') return;
      const cfg = await GetConfig() as any;
      if (cfg?.autoStartLLM === false) return;
      const status = await GetLLMStatus() as any;
      if (status?.healthy || status?.running) return;
      await StartLLMServer();
      llmState = 'loading';
    } catch {
      /* kullanıcı ayarlardan başlatabilir */
    }
  }

  function formatMonitorBytes(bytes: number): string {
    if (!bytes) return '—';
    const gb = bytes / (1024 * 1024 * 1024);
    if (gb >= 1) return gb.toFixed(1) + ' GB';
    return Math.round(bytes / (1024 * 1024)) + ' MB';
  }

  function ramUsagePercent(): number {
    if (!systemMonitor?.totalRAMBytes) return 0;
    return Math.round((systemMonitor.usedRAMBytes / systemMonitor.totalRAMBytes) * 100);
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

  function handleKeydown(e: KeyboardEvent) {
    if (!editor) return;
    const ctrl = e.ctrlKey || e.metaKey;

    if (ctrl && e.shiftKey && e.key === 'I') {
      e.preventDefault();
      requestImprovement(editor, { scope: 'paragraph', llmState });
    } else if (ctrl && e.key === 'Enter') {
      const active = reviewStore.activeReviewId;
      if (active) {
        e.preventDefault();
        handleAccept(active);
      }
    } else if (e.key === 'Escape') {
      const active = reviewStore.activeReviewId;
      if (active) {
        e.preventDefault();
        handleReject(active);
      }
    } else if (ctrl && e.shiftKey && e.key === ']') {
      e.preventDefault();
      const next = reviewStore.nextPending;
      if (next) scrollToReview(editor, next.id);
    } else if (ctrl && e.shiftKey && e.key === '[') {
      e.preventDefault();
      const prev = reviewStore.prevPending;
      if (prev) scrollToReview(editor, prev.id);
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
    pollSystemMonitor();
    statusPollTimer = setInterval(pollLLMStatus, 5000);
    monitorTimer = setInterval(pollSystemMonitor, 3000);

    if (setupInfo?.status === 'ready' || wizardSkipped) {
      tryAutoStartLLM();
    }
  });

  onDestroy(() => {
    if (statusPollTimer) clearInterval(statusPollTimer);
    if (monitorTimer) clearInterval(monitorTimer);
  });

  async function handleSetupComplete() {
    wizardSkipped = false;
    setupInfo = await CheckSetupStatus();
    tryAutoStartLLM();
  }

  function handleSetupSkip() {
    wizardSkipped = true;
  }

  function handleEditorReady(e: TipTapEditor) {
    editor = e;
  }

  function handleAccept(reviewId: string) {
    if (editor) acceptReview(editor, reviewId);
  }

  function handleReject(reviewId: string) {
    if (editor) rejectReview(editor, reviewId);
  }

  function handleAcceptDiff(reviewId: string, diffIndex: number) {
    reviewStore.acceptDiffItem(reviewId, diffIndex);
    if (editor) updateGranularDecorations(editor, reviewId);
  }

  function handleRejectDiff(reviewId: string, diffIndex: number) {
    reviewStore.rejectDiffItem(reviewId, diffIndex);
    if (editor) updateGranularDecorations(editor, reviewId);
  }

  let hasReviews = $derived(reviewStore.reviews.length > 0);
</script>

<svelte:window onkeydown={handleKeydown} />

{#if loading}
  <div class="flex items-center justify-center h-screen w-screen bg-surface">
    <div class="text-sm text-text-secondary">Yükleniyor...</div>
  </div>
{:else if setupInfo?.status !== 'ready' && !wizardSkipped}
  <SetupWizard {setupInfo} onComplete={handleSetupComplete} onSkip={handleSetupSkip} />
{:else}
  <div class="flex flex-col h-screen w-screen">
    <header class="flex items-center justify-between px-4 py-1.5 border-b border-border bg-surface shrink-0"
      style="--wails-draggable: drag;">
      <div class="flex items-center gap-3">
        <h1 class="text-sm font-semibold text-text-primary tracking-tight">Katip</h1>
        <button
          class="flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] transition-colors hover:bg-surface-secondary"
          onclick={() => showSettings = true}
          title={llmState === 'ready' ? 'AI Sunucusu hazır' : llmState === 'loading' ? 'AI Sunucusu yükleniyor...' : wizardSkipped ? 'Kurulum tamamlanmadı — Ayarlardan kuruluma devam edebilirsiniz' : 'AI Sunucusu kapalı — Ayarlardan başlatın'}
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
        <div class="flex items-center p-0.5 rounded border border-border bg-surface-secondary">
          <button
            class="text-xs px-2.5 py-1 rounded transition-colors {activePanel === 'reviews' ? 'bg-surface shadow text-primary font-medium' : 'text-text-secondary hover:text-text-primary'}"
            onclick={() => activePanel = activePanel === 'reviews' ? 'closed' : 'reviews'}
            title="AI Metin Düzeltmeleri"
          >
            Düzeltmeler{hasReviews ? ` (${reviewStore.reviews.length})` : ''}
          </button>
          <button
            class="text-xs px-2.5 py-1 rounded transition-colors {activePanel === 'comments' ? 'bg-surface shadow text-primary font-medium' : 'text-text-secondary hover:text-text-primary'}"
            onclick={() => activePanel = activePanel === 'comments' ? 'closed' : 'comments'}
            title="Kullanıcı Yorumları"
          >
            Yorumlar{commentStore.threads.filter(t => !t.resolved).length > 0 ? ` (${commentStore.threads.filter(t => !t.resolved).length})` : ''}
          </button>
          <button
            class="text-xs px-2.5 py-1 rounded transition-colors {activePanel === 'summary' ? 'bg-surface shadow text-primary font-medium' : 'text-text-secondary hover:text-text-primary'}"
            onclick={() => activePanel = activePanel === 'summary' ? 'closed' : 'summary'}
            title="Bağlam ve Hikaye Özeti"
          >
            Özet
          </button>
        </div>
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
      <Toolbar {editor} {llmState} />
    {/if}

    <div class="flex flex-1 overflow-hidden">
      <main class="flex-1 overflow-y-auto bg-surface" style="font-size: {settingsStore.fontSize}px; font-family: {settingsStore.fontFamily};">
        <div class="max-w-3xl mx-auto px-20 py-8 editor-root min-h-full">
          <Editor onReady={handleEditorReady} {llmState} />
        </div>
      </main>

      {#if editor && activePanel === 'reviews'}
        <aside class="w-72 border-l border-border bg-surface overflow-y-auto shrink-0">
          <ReviewPanel
            {editor}
            onAccept={handleAccept}
            onReject={handleReject}
            onAcceptDiff={handleAcceptDiff}
            onRejectDiff={handleRejectDiff}
          />
        </aside>
      {:else if editor && activePanel === 'comments'}
        <aside class="w-72 border-l border-border bg-surface overflow-hidden shrink-0">
          <CommentPanel {editor} />
        </aside>
      {:else if editor && activePanel === 'summary'}
        <aside class="w-72 border-l border-border bg-surface overflow-hidden shrink-0">
          <SummaryPanel {editor} {llmState} />
        </aside>
      {/if}
    </div>

    <footer class="flex items-center justify-between px-4 py-1.5 border-t border-border bg-surface text-[10px] text-text-secondary shrink-0 z-10">
      <div class="flex items-center gap-4">
        {#if editor}
          <span class="font-medium tracking-wide">{editor.storage.characterCount.words()} kelime</span>
          <span class="opacity-75">{editor.storage.characterCount.characters()} karakter</span>
        {/if}
        {#if systemMonitor}
          <span class="opacity-75 hidden sm:inline" title="RAM kullanımı">
            RAM {ramUsagePercent()}% ({formatMonitorBytes(systemMonitor.availableRAMBytes)} boş)
          </span>
          {#if systemMonitor.llmRunning}
            <span class="text-amber-600">AI aktif</span>
          {/if}
        {/if}
      </div>
      <div>Katip Yazım Motoru v0.0.1</div>
    </footer>
  </div>

  <SettingsDialog open={showSettings} onClose={() => showSettings = false} />
  <Toast />
{/if}
