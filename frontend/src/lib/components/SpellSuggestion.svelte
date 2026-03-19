<script lang="ts">
  import { getSuggestions, addWord } from '../editor/spellChecker';

  interface Props {
    word: string;
    x: number;
    y: number;
    onReplace: (newWord: string) => void;
    onClose: () => void;
  }

  let { word, x, y, onReplace, onClose }: Props = $props();

  let suggestions = $derived(getSuggestions(word, 5));

  function handleReplace(suggestion: string) {
    onReplace(suggestion);
  }

  function handleAddToDictionary() {
    addWord(word);
    onClose();
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="spell-popup"
  style="left: {x}px; top: {y}px;"
  onmousedown={(e) => e.preventDefault()}
>
  {#if suggestions.length > 0}
    <div class="spell-popup-section">
      {#each suggestions as suggestion}
        <button
          class="spell-popup-item"
          onclick={() => handleReplace(suggestion)}
        >
          {suggestion}
        </button>
      {/each}
    </div>
    <div class="spell-popup-divider"></div>
  {:else}
    <div class="spell-popup-empty">Öneri bulunamadı</div>
    <div class="spell-popup-divider"></div>
  {/if}

  <button
    class="spell-popup-item spell-popup-action"
    onclick={handleAddToDictionary}
  >
    Sözlüğe ekle
  </button>
  <button
    class="spell-popup-item spell-popup-action"
    onclick={onClose}
  >
    Yoksay
  </button>
</div>
