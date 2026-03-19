<script lang="ts">
  import type { Editor } from '@tiptap/core';
  import { reviewStore } from '../stores/reviewStore.svelte.ts';
  import ReviewCard from './ReviewCard.svelte';

  interface Props {
    editor: Editor | null;
    onAccept: (id: string) => void;
    onReject: (id: string) => void;
  }

  let { editor, onAccept, onReject }: Props = $props();

  let pendingCount = $derived(reviewStore.pendingReviews.length);
</script>

<div class="px-3 py-4">
  <div class="flex items-center justify-between mb-4 px-2">
    <span class="review-panel-header">Düzeltmeler</span>
    {#if pendingCount > 0}
      <span class="text-[10px] font-medium text-primary bg-primary/10 px-1.5 py-0.5 rounded">
        {pendingCount}
      </span>
    {/if}
  </div>

  {#if reviewStore.reviews.length === 0}
    <div class="text-center py-12 px-4">
      <p class="text-sm text-text-secondary">
        Henüz düzeltme yok.
      </p>
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
        <ReviewCard
          {review}
          onAccept={onAccept}
          onReject={onReject}
        />
      {/each}
    </div>
  {/if}
</div>
