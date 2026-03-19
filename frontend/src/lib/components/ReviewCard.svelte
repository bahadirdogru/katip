<script lang="ts">
  import type { Review } from '../stores/reviewStore.svelte.ts';

  interface Props {
    review: Review;
    onAccept: (id: string) => void;
    onReject: (id: string) => void;
  }

  let { review, onAccept, onReject }: Props = $props();

  let accentClass = $derived(
    review.status === 'accepted'
      ? 'review-card-accent-accepted'
      : review.status === 'rejected'
        ? 'review-card-accent-rejected'
        : 'review-card-accent-pending'
  );

  let timeAgo = $derived(formatTimeAgo(review.id));

  function formatTimeAgo(id: string): string {
    const ts = parseInt(id.split('-')[1] || '0', 10);
    if (!ts) return '';
    const diff = Date.now() - ts;
    const secs = Math.floor(diff / 1000);
    if (secs < 60) return 'az önce';
    const mins = Math.floor(secs / 60);
    if (mins < 60) return `${mins} dk önce`;
    const hours = Math.floor(mins / 60);
    return `${hours} sa önce`;
  }

  function truncatedDiffs() {
    const result: Array<{ type: string; text: string }> = [];
    let contextBudget = 30;

    for (const d of review.diffs) {
      if (d.type === 'equal') {
        if (d.text.length > contextBudget) {
          const start = d.text.slice(0, Math.min(12, contextBudget));
          const end = d.text.slice(-Math.min(12, contextBudget));
          result.push({ type: 'equal', text: `${start.trim()}...${end.trim()}` });
          contextBudget = 0;
        } else {
          result.push(d);
          contextBudget -= d.text.length;
        }
      } else {
        result.push(d);
      }
    }
    return result;
  }
</script>

<div class="review-card group" class:opacity-50={review.status === 'rejected'}>
  <div class="review-card-accent {accentClass}"></div>

  <div class="pl-3">
    <div class="flex items-center justify-between mb-1.5">
      <span class="text-xs font-medium text-text-secondary">{review.summary}</span>
      <span class="text-[10px] text-text-secondary/60">{timeAgo}</span>
    </div>

    <div class="text-[13px] leading-relaxed text-text-primary/90">
      {#each truncatedDiffs() as diff}
        {#if diff.type === 'equal'}
          <span class="text-text-secondary/70">{diff.text}</span>
        {:else if diff.type === 'delete'}
          <span class="diff-delete">{diff.text}</span>
        {:else if diff.type === 'insert'}
          <span class="diff-insert">{diff.text}</span>
        {/if}
      {/each}
    </div>

    {#if review.status === 'pending'}
      <div class="review-card-actions flex items-center gap-1 mt-2 justify-end">
        <button
          class="review-action-btn review-action-btn-accept"
          onclick={() => onAccept(review.id)}
          title="Onayla"
        >
          ✓
        </button>
        <button
          class="review-action-btn review-action-btn-reject"
          onclick={() => onReject(review.id)}
          title="Reddet"
        >
          ✕
        </button>
      </div>
    {:else}
      <div class="mt-1.5">
        <span class="text-[10px] font-medium {review.status === 'accepted' ? 'text-green-600' : 'text-text-secondary/50'}">
          {review.status === 'accepted' ? 'Onaylandı' : 'Reddedildi'}
        </span>
      </div>
    {/if}
  </div>
</div>
