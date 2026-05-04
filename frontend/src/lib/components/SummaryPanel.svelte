<script lang="ts">
  import { documentStore } from '../stores/documentStore.svelte.ts';
  import { GeneratePlotSummary } from '../../../bindings/katip/internal/service/katipservice.js';
  import type { Editor } from '@tiptap/core';

  interface Props {
    editor: Editor;
  }

  let { editor }: Props = $props();
  let generating = $state(false);

  async function handleAutoSummarize() {
    if (generating) return;
    const fullText = editor.getText().trim();
    if (fullText.length < 50) {
      alert("Metin özet çıkarmak için çok kısa! Önce hikayenizi yazmalısınız.");
      return;
    }
    
    generating = true;
    try {
      const summary = await GeneratePlotSummary(fullText);
      if (summary) {
        documentStore.setPlotSummary(summary);
      }
    } catch (err: any) {
      console.error("Özet çıkarma hatası:", err);
      alert("AI bağlantı hatası: " + err);
    } finally {
      generating = false;
    }
  }
</script>

<div class="h-full flex flex-col pt-3 bg-surface border-l border-border">
  <div class="px-4 pb-3 flex items-center justify-between border-b border-border">
    <h2 class="text-xs font-semibold text-text-primary tracking-wide uppercase">Hikaye Özeti (Bağlam)</h2>
  </div>

  <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-3">
    <p class="text-[11px] text-text-secondary leading-relaxed">
      Yapay zekanın (LLM) düzeltme yaparken karakter kopukluğu yaşamaması veya yazarın tarzını unutmaması için buraya bir "olay örgüsü" ya da "karakter notları" yazabilirsiniz. AI her düzeltmede bu metni referans alacaktır.
    </p>

    <button 
      class="w-full py-1.5 mt-1 text-[11px] font-medium rounded-md flex items-center justify-center gap-1.5 transition-colors {generating ? 'bg-surface-secondary text-text-secondary cursor-wait disabled:opacity-50' : 'bg-primary/10 text-primary border border-primary/20 hover:bg-primary hover:text-white'}"
      onclick={handleAutoSummarize}
      disabled={generating}
    >
      {#if generating}
        <span class="animate-spin text-sm">⏳</span> Analiz Ediliyor...
      {:else}
        <span class="text-sm">✨</span> Otomatik Bağlam Çıkar
      {/if}
    </button>

    <textarea
      class="flex-1 w-full bg-surface-secondary border border-border rounded-lg p-3 text-[12px] text-text-primary focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/30 resize-none transition-shadow"
      placeholder="Örn: Hikaye 19. yüzyıl Londra'sında geçiyor. Baş karakter Arthur, şüpheci ve zeki bir dedektiftir..."
      value={documentStore.plotSummary}
      oninput={(e) => documentStore.setPlotSummary(e.currentTarget.value)}
    ></textarea>
  </div>
</div>
