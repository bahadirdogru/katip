<script lang="ts">
  import { toastStore } from '../stores/toastStore.svelte.ts';

  const typeStyles: Record<string, string> = {
    info: 'bg-surface border-border text-text-primary',
    success: 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200',
    warning: 'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200',
    error: 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200',
  };
</script>

<div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-[100] flex flex-col gap-2 items-center pointer-events-none">
  {#each toastStore.toasts as toast (toast.id)}
    <div
      class="pointer-events-auto px-4 py-2.5 rounded-lg border shadow-lg text-sm max-w-md animate-[slideUp_0.2s_ease] {typeStyles[toast.type]}"
      role="alert"
    >
      <div class="flex items-center gap-2">
        <span class="flex-1">{toast.message}</span>
        <button
          class="text-text-secondary hover:text-text-primary opacity-60 hover:opacity-100 text-xs"
          onclick={() => toastStore.dismiss(toast.id)}
        >✕</button>
      </div>
    </div>
  {/each}
</div>

<style>
  @keyframes slideUp {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
