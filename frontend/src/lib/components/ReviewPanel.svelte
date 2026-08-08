<script lang="ts">
  import type { Editor } from '@tiptap/core';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import ReviewCard from './ReviewCard.svelte';
  import {
    scrollToReview,
    acceptAllPending,
    rejectAllPending,
  } from '../editor/aiReviewService.ts';

  interface Props {
    editor: Editor | null;
    onAccept: (id: string) => void;
    onReject: (id: string) => void;
    onAcceptDiff?: (reviewId: string, diffIndex: number) => void;
    onRejectDiff?: (reviewId: string, diffIndex: number) => void;
  }

  let { editor, onAccept, onReject, onAcceptDiff, onRejectDiff }: Props = $props();

  let pendingCount = $derived(reviewStore.pendingReviews.length);
  let hasCompleted = $derived(reviewStore.reviews.some(r => r.status === 'accepted' || r.status === 'rejected'));

  function handleNavigateNext() {
    const next = reviewStore.nextPending;
    if (next && editor) scrollToReview(editor, next.id);
  }

  function handleNavigatePrev() {
    const prev = reviewStore.prevPending;
    if (prev && editor) scrollToReview(editor, prev.id);
  }

  function handleCardClick(reviewId: string) {
    if (editor) scrollToReview(editor, reviewId);
  }

  function handleAcceptAll() {
    if (editor) acceptAllPending(editor);
  }

  function handleRejectAll() {
    if (editor) rejectAllPending(editor);
  }

  function handleClearCompleted() {
    reviewStore.clearCompleted();
  }
</script>

<div class="px-3 py-4">
  <div class="flex items-center justify-between mb-3 px-2">
    <span class="review-panel-header">Düzeltmeler</span>
    {#if pendingCount > 0}
      <span class="text-[10px] font-medium text-primary bg-primary/10 px-1.5 py-0.5 rounded">
        {pendingCount}
      </span>
    {/if}
  </div>

  {#if pendingCount > 0}
    <div class="flex flex-wrap gap-1 mb-3 px-2">
      <button
        class="text-[10px] px-2 py-1 rounded bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors"
        onclick={handleAcceptAll}
      >Tümünü Onayla</button>
      <button
        class="text-[10px] px-2 py-1 rounded bg-surface-secondary text-text-secondary hover:bg-surface border border-border transition-colors"
        onclick={handleRejectAll}
      >Tümünü Reddet</button>
      <button
        class="text-[10px] px-2 py-1 rounded text-text-secondary hover:bg-surface-secondary transition-colors"
        onclick={handleNavigatePrev}
        title="Önceki öneri"
      >←</button>
      <button
        class="text-[10px] px-2 py-1 rounded text-text-secondary hover:bg-surface-secondary transition-colors"
        onclick={handleNavigateNext}
        title="Sonraki öneri"
      >→</button>
    </div>
  {/if}

  {#if hasCompleted}
    <div class="px-2 mb-3">
      <button
        class="text-[10px] text-text-secondary hover:text-text-primary underline"
        onclick={handleClearCompleted}
      >Tamamlananları Temizle</button>
    </div>
  {/if}

  {#if reviewStore.reviews.length === 0}
    <div class="text-center py-12 px-4">
      <p class="text-sm text-text-secondary">Henüz düzeltme yok.</p>
      <p class="text-xs text-text-secondary/60 mt-2 leading-relaxed">
        Bir paragrafı seçip "AI İyileştir" butonuna tıklayın.
      </p>
    </div>
  {:else}
    <div class="flex flex-col">
      {#each reviewStore.reviews as review, i (review.id)}
        {#if i > 0}
          <div class="review-separator"></div>
        {/if}
        <div
          class="cursor-pointer rounded-md transition-colors {reviewStore.activeReviewId === review.id ? 'ring-1 ring-primary/30 bg-primary/5' : ''}"
          onclick={() => handleCardClick(review.id)}
          role="button"
          tabindex="0"
          onkeydown={(e) => e.key === 'Enter' && handleCardClick(review.id)}
        >
          <ReviewCard
            {review}
            onAccept={onAccept}
            onReject={onReject}
            {onAcceptDiff}
            {onRejectDiff}
          />
        </div>
      {/each}
    </div>
  {/if}
</div>
