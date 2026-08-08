import type { Editor } from '@tiptap/core';
import {
  ImproveParagraph,
  ImproveSelection,
  CancelImprovement,
  GetImproveStreamProgress,
} from '../../../bindings/katip/internal/service/katipservice.js';
import { reviewStore, type ImprovementMode, type ReviewKind } from '../stores/reviewStore.svelte.ts';
import { documentStore } from '../stores/documentStore.svelte.ts';
import { toastStore } from '../stores/toastStore.svelte.ts';
import { versionHistoryStore } from '../stores/versionHistoryStore.svelte.ts';
import {
  applyDiffForReview,
  clearReviewDecorations,
  updateReviewDecorations,
  type DiffRange,
} from './diffDecorations.ts';
import type { DiffItem } from '../stores/reviewStore.svelte.ts';

export type LLMState = 'off' | 'loading' | 'ready';
export type ImproveScope = 'paragraph' | 'selection';

export interface ImproveOptions {
  scope?: ImproveScope;
  mode?: ImprovementMode;
  kind?: ReviewKind;
  llmState?: LLMState;
  onStreamProgress?: (partial: string) => void;
}

let activeAbort = false;

export function findTextblockAtPos(
  editor: Editor,
  paragraphPos: number
): { node: any; pos: number; from: number; text: string } | null {
  const { state } = editor.view;
  const node = state.doc.nodeAt(paragraphPos);
  if (node?.isTextblock) {
    return { node, pos: paragraphPos, from: paragraphPos + 1, text: node.textContent };
  }
  return null;
}

export function findTextblockByText(
  editor: Editor,
  originalText: string,
  hintPos?: number
): { from: number; pos: number; text: string } | null {
  const { state } = editor.view;
  let best: { from: number; pos: number; text: string } | null = null;
  let bestDist = Infinity;

  state.doc.descendants((node, pos) => {
    if (node.isTextblock) {
      const text = node.textContent;
      if (text === originalText || text.trim() === originalText.trim()) {
        const dist = hintPos !== undefined ? Math.abs(pos - hintPos) : 0;
        if (dist < bestDist) {
          bestDist = dist;
          best = { from: pos + 1, pos, text };
        }
      }
    }
  });

  return best;
}

export function collectParagraphs(
  editor: Editor,
  minLength = 10
): { id: string; text: string; pos: number }[] {
  const paragraphs: { id: string; text: string; pos: number }[] = [];
  editor.view.state.doc.descendants((node, pos) => {
    if (node.isTextblock) {
      const text = node.textContent.trim();
      if (text.length >= minLength) {
        paragraphs.push({ id: `p-${pos}`, text, pos });
      }
    }
  });
  return paragraphs;
}

function guardLLM(llmState?: LLMState): boolean {
  if (!llmState || llmState === 'ready') return true;
  if (llmState === 'loading') {
    toastStore.warning('AI sunucusu yükleniyor, lütfen bekleyin...');
  } else {
    toastStore.error('AI sunucusu çalışmıyor. Ayarlardan başlatın.');
  }
  return false;
}

function mapDiffs(diffs: any[]): DiffItem[] {
  return diffs.map((d) => ({
    type: d.type as 'equal' | 'insert' | 'delete',
    text: d.text,
  }));
}

function handleImproveResult(
  editor: Editor,
  result: any,
  paragraphPos: number,
  mode?: ImprovementMode,
  kind?: ReviewKind,
  selectionFrom?: number,
  selectionTo?: number
): string | null {
  if (!result?.diffs?.length) {
    toastStore.info('Bu metinde düzeltme önerilmedi.');
    return null;
  }

  const hasChanges = result.diffs.some((d: any) => d.type !== 'equal');
  if (!hasChanges) {
    toastStore.info('Bu metinde düzeltme önerilmedi.');
    return null;
  }

  const mappedDiffs = mapDiffs(result.diffs);
  const reviewId = reviewStore.addReview({
    paragraphId: result.paragraphId,
    paragraphPos,
    selectionFrom,
    selectionTo,
    summary: result.summary,
    original: result.original,
    improved: result.improved,
    diffs: mappedDiffs,
    mode,
    kind,
    changeType: result.changeType,
    ruleName: result.ruleName,
  });

  const block = findTextblockAtPos(editor, paragraphPos);
  if (block) {
    applyDiffForReview(editor, reviewId, result.paragraphId, block.from, mappedDiffs);
  }

  return reviewId;
}

export async function requestImprovement(
  editor: Editor,
  options: ImproveOptions = {}
): Promise<string | null> {
  const { scope = 'paragraph', mode = 'fix', kind = 'improvement', llmState } = options;
  if (!guardLLM(llmState)) return null;

  const { state } = editor.view;
  const plotSummary = documentStore.plotSummary;

  try {
    if (scope === 'selection') {
      const { from, to, empty } = state.selection;
      if (empty || to - from < 3) {
        toastStore.warning('Lütfen iyileştirilecek metni seçin.');
        return null;
      }
      const text = state.doc.textBetween(from, to, ' ');
      const resolved = state.selection.$from;
      const paragraphPos = resolved.before(resolved.depth);

      const result = await ImproveSelection(
        `sel-${from}`,
        text,
        plotSummary,
        mode
      );
      return handleImproveResult(editor, result, paragraphPos, mode, kind, from, to);
    }

    const resolved = state.selection.$from;
    const node = resolved.parent;
    if (!node.isTextblock || node.textContent.trim().length < 3) {
      toastStore.warning('İmleci bir paragrafın içine yerleştirin.');
      return null;
    }

    const paragraphPos = resolved.before(resolved.depth);
    const text = node.textContent.trim();

    const result = await ImproveParagraph(
      `p-${paragraphPos}`,
      text,
      plotSummary,
      mode
    );
    return handleImproveResult(editor, result, paragraphPos, mode, kind);
  } catch (err: any) {
    const msg = String(err?.message || err);
    if (msg.includes('timeout') || msg.includes('bağlanılamadı')) {
      toastStore.error('AI yanıt vermedi. Paragrafı kısaltmayı veya modeli kontrol etmeyi deneyin.');
    } else if (msg.includes('iptal')) {
      toastStore.info('AI işlemi iptal edildi.');
    } else {
      toastStore.error('AI iyileştirme hatası: ' + msg);
    }
    return null;
  }
}

export async function requestImprovementWithStream(
  editor: Editor,
  options: ImproveOptions = {}
): Promise<string | null> {
  activeAbort = false;
  const reviewId = await requestImprovement(editor, options);
  return reviewId;
}

export function cancelActiveImprovement() {
  activeAbort = true;
  CancelImprovement().catch(() => {});
}

export function acceptReview(editor: Editor, reviewId: string): boolean {
  const review = reviewStore.reviews.find((r) => r.id === reviewId);
  if (!review || !editor) return false;

  clearReviewDecorations(editor, reviewId);

  const improvedText =
    review.status === 'partial'
      ? reviewStore.buildPartialImproved(review)
      : review.improved;

  const { state, dispatch } = editor.view;
  let found = false;

  if (review.selectionFrom !== undefined && review.selectionTo !== undefined) {
    const tr = state.tr.replaceWith(
      review.selectionFrom,
      review.selectionTo,
      state.schema.text(improvedText)
    );
    dispatch(tr);
    found = true;
  } else {
    const block = findTextblockByText(editor, review.original, review.paragraphPos);
    if (block) {
      const tr = state.tr.replaceWith(
        block.from,
        block.from + block.text.length,
        state.schema.text(improvedText)
      );
      dispatch(tr);
      found = true;
    }
  }

  if (found) {
    reviewStore.acceptReview(reviewId);
    versionHistoryStore.addEntry({
      paragraphId: review.paragraphId,
      original: review.original,
      improved: improvedText,
      summary: review.summary,
      mode: review.mode,
    });
  } else {
    toastStore.warning('Paragraf bulunamadı; metin değiştirilmiş olabilir.');
  }

  return found;
}

export function rejectReview(editor: Editor, reviewId: string) {
  clearReviewDecorations(editor, reviewId);
  reviewStore.rejectReview(reviewId);
}

export function acceptAllPending(editor: Editor) {
  const pending = [...reviewStore.pendingReviews];
  for (const r of pending) {
    acceptReview(editor, r.id);
  }
  toastStore.success(`${pending.length} öneri onaylandı.`);
}

export function rejectAllPending(editor: Editor) {
  const pending = [...reviewStore.pendingReviews];
  for (const r of pending) {
    rejectReview(editor, r.id);
  }
  toastStore.info(`${pending.length} öneri reddedildi.`);
}

export function scrollToReview(editor: Editor, reviewId: string) {
  const review = reviewStore.reviews.find((r) => r.id === reviewId);
  if (!review) return;

  reviewStore.setActiveReview(reviewId);
  const block = findTextblockByText(editor, review.original, review.paragraphPos);
  if (!block) return;

  const dom = editor.view.domAtPos(block.pos);
  const el = (dom.node as HTMLElement).nodeType === 1
    ? (dom.node as HTMLElement)
    : (dom.node as HTMLElement).parentElement;

  el?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  el?.classList.add('review-highlight-pulse');
  setTimeout(() => el?.classList.remove('review-highlight-pulse'), 1500);
}

export function updateGranularDecorations(editor: Editor, reviewId: string) {
  const review = reviewStore.reviews.find((r) => r.id === reviewId);
  if (!review) return;

  const block = findTextblockByText(editor, review.original, review.paragraphPos);
  if (!block) return;

  updateReviewDecorations(editor, {
    reviewId,
    paragraphId: review.paragraphId,
    diffs: review.diffs,
    from: block.from,
    to: block.from,
  } satisfies DiffRange);
}
