<script lang="ts">
  import { documentStore } from '../stores/documentStore.svelte.ts';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import { versionHistoryStore } from '../stores/versionHistoryStore.svelte.ts';
  import { toastStore } from '../stores/toastStore.svelte.ts';
  import {
    GeneratePlotSummary,
    CheckConsistency,
    CheckTarzStyle,
    GetTarzProfiles,
  } from '../../../bindings/katip/internal/service/katipservice.js';
  import type { Editor } from '@tiptap/core';
  import type { LLMState } from '../editor/aiReviewService.ts';
  import { collectParagraphs } from '../editor/aiReviewService.ts';

  interface Props {
    editor: Editor;
    llmState?: LLMState;
  }

  let { editor, llmState = 'off' }: Props = $props();
  let generating = $state(false);
  let checkingConsistency = $state(false);
  let checkingTarz = $state(false);
  let tarzProfiles = $state<string[]>([]);
  let selectedProfile = $state('varsayilan');
  let showHistory = $state(false);

  let aiReady = $derived(llmState === 'ready');

  async function loadProfiles() {
    try {
      const profiles = await GetTarzProfiles();
      if (profiles?.length) {
        tarzProfiles = profiles;
        selectedProfile = profiles[0];
      }
    } catch {
      tarzProfiles = ['varsayilan'];
    }
  }

  $effect(() => {
    if (aiReady) loadProfiles();
  });

  async function handleAutoSummarize() {
    if (generating || !aiReady) {
      if (!aiReady) toastStore.error('AI sunucusu çalışmıyor. Ayarlardan başlatın.');
      return;
    }
    const fullText = editor.getText().trim();
    if (fullText.length < 50) {
      toastStore.warning('Metin özet çıkarmak için çok kısa! Önce hikayenizi yazmalısınız.');
      return;
    }

    generating = true;
    try {
      const summary = await GeneratePlotSummary(fullText);
      if (summary) {
        documentStore.setPlotSummary(summary);
        toastStore.success('Hikaye özeti oluşturuldu.');
      }
    } catch (err: any) {
      toastStore.error('AI bağlantı hatası: ' + (err?.message || err));
    } finally {
      generating = false;
    }
  }

  async function handleConsistencyCheck() {
    if (checkingConsistency || !aiReady) return;
    const plotSummary = documentStore.plotSummary;
    if (!plotSummary) {
      toastStore.warning('Önce hikaye özeti oluşturun veya yazın.');
      return;
    }

    const paragraphs = collectParagraphs(editor, 20);
    if (paragraphs.length === 0) return;

    checkingConsistency = true;
    let found = 0;
    try {
      for (const para of paragraphs.slice(0, 5)) {
        const result = await CheckConsistency(para.text, plotSummary);
        if (result && !result.toLowerCase().includes('tutarlı')) {
          reviewStore.addReview({
            paragraphId: para.id,
            paragraphPos: para.pos,
            summary: 'Tutarlılık uyarısı',
            original: para.text,
            improved: para.text,
            diffs: [{ type: 'equal', text: result }],
            kind: 'consistency',
            changeType: 'consistency',
          });
          found++;
        }
      }
      toastStore.info(found > 0 ? `${found} tutarlılık uyarısı bulundu.` : 'Tutarlılık sorunu bulunamadı.');
    } catch (err: any) {
      toastStore.error('Tutarlılık denetimi hatası: ' + (err?.message || err));
    } finally {
      checkingConsistency = false;
    }
  }

  async function handleTarzCheck() {
    if (checkingTarz) return;
    const paragraphs = collectParagraphs(editor, 10);
    if (paragraphs.length === 0) return;

    checkingTarz = true;
    let found = 0;
    try {
      for (const para of paragraphs) {
        const violations = await CheckTarzStyle(selectedProfile, para.text);
        if (violations?.length) {
          for (const v of violations) {
            reviewStore.addReview({
              paragraphId: para.id,
              paragraphPos: para.pos,
              summary: v.ruleName || 'Tarz kuralı',
              original: v.original,
              improved: v.suggestion,
              diffs: [
                { type: 'delete', text: v.original },
                { type: 'insert', text: v.suggestion },
              ],
              kind: 'style',
              changeType: 'style',
              ruleName: v.ruleName,
            });
            found++;
          }
        }
      }
      toastStore.info(found > 0 ? `${found} tarz uyarısı bulundu.` : 'Tarz kuralı ihlali bulunamadı.');
    } catch (err: any) {
      toastStore.error('Tarz denetimi hatası: ' + (err?.message || err));
    } finally {
      checkingTarz = false;
    }
  }
</script>

<div class="h-full flex flex-col pt-3 bg-surface border-l border-border">
  <div class="px-4 pb-3 flex items-center justify-between border-b border-border">
    <h2 class="text-xs font-semibold text-text-primary tracking-wide uppercase">Hikaye Özeti (Bağlam)</h2>
    <button
      class="text-[10px] text-text-secondary hover:text-primary"
      onclick={() => showHistory = !showHistory}
    >{showHistory ? 'Özet' : 'Geçmiş'}</button>
  </div>

  {#if showHistory}
    <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-2">
      <p class="text-[11px] text-text-secondary">Onaylanan AI değişiklikleri geçmişi</p>
      {#if versionHistoryStore.entries.length === 0}
        <p class="text-xs text-text-secondary/60 italic">Henüz geçmiş yok.</p>
      {:else}
        {#each versionHistoryStore.entries as entry (entry.id)}
          <div class="p-2 rounded border border-border bg-surface-secondary text-[11px]">
            <div class="text-text-secondary mb-1">{entry.summary}</div>
            <div class="text-text-primary/70 line-through">{entry.original.slice(0, 60)}...</div>
            <div class="text-green-600">{entry.improved.slice(0, 60)}...</div>
          </div>
        {/each}
      {/if}
    </div>
  {:else}
    <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-3">
      <p class="text-[11px] text-text-secondary leading-relaxed">
        Yapay zekanın düzeltme yaparken karakter kopukluğu yaşamaması için buraya olay örgüsü ya da karakter notları yazabilirsiniz. Verileriniz cihazınızdan çıkmaz.
      </p>

      <button
        class="w-full py-1.5 text-[11px] font-medium rounded-md flex items-center justify-center gap-1.5 transition-colors disabled:opacity-40
          {generating ? 'bg-surface-secondary text-text-secondary cursor-wait' : 'bg-primary/10 text-primary border border-primary/20 hover:bg-primary hover:text-white'}"
        onclick={handleAutoSummarize}
        disabled={generating || !aiReady}
      >
        {#if generating}
          <span class="animate-spin text-sm">⏳</span> Analiz Ediliyor...
        {:else}
          <span class="text-sm">✨</span> Otomatik Bağlam Çıkar
        {/if}
      </button>

      <button
        class="w-full py-1.5 text-[11px] font-medium rounded-md border border-border text-text-secondary hover:bg-surface-secondary disabled:opacity-40"
        onclick={handleConsistencyCheck}
        disabled={checkingConsistency || !aiReady}
      >
        {checkingConsistency ? '⏳ Kontrol ediliyor...' : '🔍 Tutarlılık Tara'}
      </button>

      <div class="flex gap-1">
        {#if tarzProfiles.length > 0}
          <select
            class="flex-1 text-[11px] bg-surface-secondary border border-border rounded px-2 py-1.5"
            bind:value={selectedProfile}
          >
            {#each tarzProfiles as profile}
              <option value={profile}>{profile}</option>
            {/each}
          </select>
        {/if}
        <button
          class="px-3 py-1.5 text-[11px] font-medium rounded-md border border-amber-200 dark:border-amber-800 text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-900/20 hover:bg-amber-100 disabled:opacity-40"
          onclick={handleTarzCheck}
          disabled={checkingTarz}
        >Tarz Denetimi</button>
      </div>

      <textarea
        class="flex-1 w-full bg-surface-secondary border border-border rounded-lg p-3 text-[12px] text-text-primary focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/30 resize-none transition-shadow"
        placeholder="Örn: Hikaye 19. yüzyıl Londra'sında geçiyor. Baş karakter Arthur, şüpheci ve zeki bir dedektiftir..."
        value={documentStore.plotSummary}
        oninput={(e) => documentStore.setPlotSummary(e.currentTarget.value)}
      ></textarea>
    </div>
  {/if}
</div>
