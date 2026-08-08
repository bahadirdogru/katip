<script lang="ts">
  import { commentStore } from '../stores/commentStore.svelte.ts';
  import type { Editor } from '@tiptap/core';

  interface Props {
    editor: Editor;
  }
  let { editor }: Props = $props();

  let activeReplyText = $state<{ [key: string]: string }>({});

  function handleReply(threadId: string) {
    const text = activeReplyText[threadId]?.trim();
    if (!text) return;
    commentStore.addReply(threadId, text);
    activeReplyText[threadId] = '';
  }

  function handleResolve(threadId: string) {
    commentStore.resolveThread(threadId);
    editor.chain().focus().unsetComment(threadId).run();
  }

  function formatDate(isoString: string) {
    const d = new Date(isoString);
    return d.toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit' }) + ' - ' + d.toLocaleDateString('tr-TR', { day: '2-digit', month: 'short' });
  }

  function scrollToComment(threadId: string) {
    commentStore.setActive(threadId);
    // Find mark in DOM
    const el = document.querySelector(`.comment-mark[data-comment-id="${threadId}"]`);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  function quoteText(str: string) {
    if (str.length > 60) return str.substring(0, 60) + '...';
    return str;
  }

  let visibleThreads = $derived(commentStore.threads.filter(t => !t.resolved));
</script>

<div class="h-full flex flex-col pt-3 bg-surface border-l border-border">
  <div class="px-4 pb-3 flex items-center justify-between border-b border-border">
    <h2 class="text-xs font-semibold text-text-primary tracking-wide uppercase">Yorumlar</h2>
    <span class="text-[10px] bg-primary/10 text-primary px-1.5 py-0.5 rounded-full font-bold">{visibleThreads.length}</span>
  </div>

  <div class="flex-1 overflow-y-auto p-3 space-y-3">
    {#if visibleThreads.length === 0}
      <div class="text-center py-8">
        <span class="text-2xl opacity-40">💬</span>
        <p class="text-xs text-text-secondary mt-2">Henüz yorum yok.</p>
        <p class="text-[10px] text-text-secondary opacity-70 mt-1">Metni seçip "Yorum Yap" butonuna tıklayarak ekleyebilirsiniz.</p>
      </div>
    {/if}

    {#each visibleThreads as thread (thread.id)}
      {@const isActive = commentStore.activeCommentId === thread.id}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div 
        class="border rounded-lg transition-all duration-200 overflow-hidden cursor-pointer
          {isActive ? 'border-primary ring-1 ring-primary/20 shadow-md transform -translate-y-0.5' : 'border-border bg-surface-secondary hover:border-gray-300'}"
        onclick={() => scrollToComment(thread.id)}
      >
        <div class="p-3">
          <!-- Quote -->
          <div class="pl-2 border-l-2 border-primary/40 mb-3 text-[11px] text-text-secondary italic line-clamp-3">
            "{quoteText(thread.quote)}"
          </div>

          <!-- Replies -->
          <div class="space-y-3 mt-2">
            {#each thread.replies as reply, idx}
              <div class="flex gap-2">
                <div class="w-5 h-5 rounded bg-gradient-to-br from-primary/80 to-primary-dark flex items-center justify-center text-[10px] text-white font-bold shrink-0">
                  {reply.author.charAt(0).toUpperCase()}
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between gap-1.5">
                    <span class="text-[10px] font-semibold text-text-primary truncate">{reply.author}</span>
                    <span class="text-[9px] text-text-secondary shrink-0">{formatDate(reply.timestamp)}</span>
                  </div>
                  <p class="text-[11px] text-text-primary mt-0.5 leading-relaxed">{reply.text}</p>
                </div>
              </div>
            {/each}
          </div>
        </div>

        <!-- Action / Reply Box -->
        {#if isActive}
          <div class="bg-surface p-2 border-t border-border">
            <div class="flex flex-col gap-2">
              <textarea 
                bind:value={activeReplyText[thread.id]} 
                placeholder="Yanıt yaz..."
                class="w-full bg-surface-secondary border border-border rounded px-2 py-1.5 text-[11px] focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 resize-none"
                rows="2"
              ></textarea>
              <div class="flex items-center justify-between">
                <button 
                  class="text-[10px] text-text-secondary hover:text-green-600 transition-colors flex items-center gap-1"
                  onclick={(e) => { e.stopPropagation(); handleResolve(thread.id); }}
                >
                  <span>✓</span> Çözüldü yap
                </button>
                <button 
                  class="bg-primary hover:bg-primary-dark text-white text-[10px] px-3 py-1 rounded font-medium transition-colors disabled:opacity-50"
                  disabled={!activeReplyText[thread.id]?.trim()}
                  onclick={(e) => { e.stopPropagation(); handleReply(thread.id); }}
                >
                  Yanıtla
                </button>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>
