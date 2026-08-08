<script lang="ts">
  import type { Review } from '../stores/reviewStore.svelte.ts';

  interface Props {
    review: Review;
    onAccept: (id: string) => void;
    onReject: (id: string) => void;
    onAcceptDiff?: (reviewId: string, diffIndex: number) => void;
    onRejectDiff?: (reviewId: string, diffIndex: number) => void;
  }

  let { review, onAccept, onReject, onAcceptDiff, onRejectDiff }: Props = $props();

  let accentClass = $derived(
    review.status === 'accepted'
      ? 'review-card-accent-accepted'
      : review.status === 'rejected'
        ? 'review-card-accent-rejected'
        : review.kind === 'consistency'
          ? 'review-card-accent-warning'
          : review.kind === 'style'
            ? 'review-card-accent-style'
            : 'review-card-accent-pending'
  );

  let timeAgo = $derived(formatTimeAgo(review.id));

  const changeTypeLabels: Record<string, string> = {
    spelling: 'Yazım',
    grammar: 'Anlatım',
    style: 'Stil',
    consistency: 'Tutarlılık',
  };

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
    const result: Array<{ type: string; text: string; index: number; rejected?: boolean; accepted?: boolean }> = [];
    let contextBudget = 30;

    review.diffs.forEach((d, index) => {
      if (d.type === 'equal') {
        if (d.text.length > contextBudget) {
          const start = d.text.slice(0, Math.min(12, contextBudget));
          const end = d.text.slice(-Math.min(12, contextBudget));
          result.push({ type: 'equal', text: `${start.trim()}...${end.trim()}`, index });
          contextBudget = 0;
        } else {
          result.push({ ...d, index });
          contextBudget -= d.text.length;
        }
      } else {
        result.push({ ...d, index });
      }
    });
    return result;
  }

  function buildDiffSegments() {
    type Segment =
      | { kind: 'equal'; text: string }
      | { kind: 'change'; deleteText: string; insertText: string; indices: number[]; rejected?: boolean; accepted?: boolean };

    const segments: Segment[] = [];
    const diffs = truncatedDiffs();
    let i = 0;

    while (i < diffs.length) {
      const current = diffs[i];
      if (current.type === 'equal') {
        segments.push({ kind: 'equal', text: current.text });
        i++;
        continue;
      }

      const next = diffs[i + 1];
      if (current.type === 'delete' && next?.type === 'insert') {
        segments.push({
          kind: 'change',
          deleteText: current.text,
          insertText: next.text,
          indices: [current.index, next.index],
          rejected: current.rejected || next.rejected,
          accepted: current.accepted && next.accepted,
        });
        i += 2;
        continue;
      }

      segments.push({
        kind: 'change',
        deleteText: current.type === 'delete' ? current.text : '',
        insertText: current.type === 'insert' ? current.text : '',
        indices: [current.index],
        rejected: current.rejected,
        accepted: current.accepted,
      });
      i++;
    }

    return segments;
  }

  function handleGroupAccept(e: MouseEvent, indices: number[]) {
    e.stopPropagation();
    for (const index of indices) onAcceptDiff?.(review.id, index);
  }

  function handleGroupReject(e: MouseEvent, indices: number[]) {
    e.stopPropagation();
    for (const index of indices) onRejectDiff?.(review.id, index);
  }

  let diffSegments = $derived(buildDiffSegments());
  let changeRows = $derived(diffSegments.filter((s): s is Extract<typeof s, { kind: 'change' }> => s.kind === 'change'));

  function rowStateClass(segment: { rejected?: boolean; accepted?: boolean }): string {
    if (segment.accepted) return 'review-diff-row-accepted';
    if (segment.rejected) return 'review-diff-row-rejected';
    return '';
  }

  function handleAcceptClick(e: MouseEvent) {
    e.stopPropagation();
    onAccept(review.id);
  }

  function handleRejectClick(e: MouseEvent) {
    e.stopPropagation();
    onReject(review.id);
  }
</script>

<div class="review-card group" class:opacity-50={review.status === 'rejected'}>
  <div class="review-card-accent {accentClass}"></div>

  <div class="pl-3">
    <div class="flex items-center justify-between mb-1.5">
      <div class="flex items-center gap-1.5">
        <span class="text-xs font-medium text-text-secondary">{review.summary}</span>
        {#if review.changeType}
          <span class="text-[9px] px-1 py-0.5 rounded bg-surface-secondary text-text-secondary/70">
            {changeTypeLabels[review.changeType] ?? review.changeType}
          </span>
        {/if}
        <span class="text-[9px] text-text-secondary/50">AI önerisi</span>
      </div>
      <span class="text-[10px] text-text-secondary/60">{timeAgo}</span>
    </div>

    {#if changeRows.length > 0 && (review.status === 'pending' || review.status === 'partial')}
      <div class="review-diff-list">
        {#each changeRows as segment, rowIdx}
          <div class="review-diff-row {rowStateClass(segment)}">
            <span class="review-diff-row-num">{rowIdx + 1}</span>
            <div class="review-diff-row-body min-w-0 flex-1">
              {#if segment.deleteText}
                <span class="diff-delete text-[12px]">{segment.deleteText}</span>
              {/if}
              {#if segment.deleteText && segment.insertText}
                <span class="text-text-secondary/40 mx-1 text-[10px]">→</span>
              {/if}
              {#if segment.insertText}
                <span class="diff-insert text-[12px]">{segment.insertText}</span>
              {/if}
            </div>
            <div class="review-diff-row-actions">
              <button
                class="review-diff-btn review-diff-btn-accept"
                onclick={(e) => handleGroupAccept(e, segment.indices)}
                title="Onayla"
              >✓</button>
              <button
                class="review-diff-btn review-diff-btn-reject"
                onclick={(e) => handleGroupReject(e, segment.indices)}
                title="Reddet"
              >✕</button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="text-[13px] leading-relaxed text-text-primary/90">
        {#each diffSegments as segment}
          {#if segment.kind === 'equal'}
            <span class="text-text-secondary/70">{segment.text}</span>
          {:else}
            <span class="{segment.rejected ? 'opacity-50' : ''}">
              {#if segment.deleteText}
                <span class="diff-delete">{segment.deleteText}</span>
              {/if}
              {#if segment.insertText}
                <span class="diff-insert">{segment.insertText}</span>
              {/if}
            </span>
          {/if}
        {/each}
      </div>
    {/if}

    {#if review.status === 'pending' || review.status === 'partial'}
      <div class="review-card-actions flex items-center gap-1 mt-2 justify-end">
        <button
          class="review-action-btn review-action-btn-accept"
          onclick={handleAcceptClick}
          title="Onayla"
        >✓</button>
        <button
          class="review-action-btn review-action-btn-reject"
          onclick={handleRejectClick}
          title="Reddet"
        >✕</button>
      </div>
    {:else}
      <div class="mt-1.5">
        <span class="text-[10px] font-medium {review.status === 'accepted' ? 'text-green-600' : 'text-text-secondary/50'}">
          {review.status === 'accepted' ? 'Onaylandı' : review.status === 'partial' ? 'Kısmen onaylandı' : 'Reddedildi'}
        </span>
      </div>
    {/if}
  </div>
</div>
