import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';
import type { DiffItem } from '../stores/reviewStore.svelte.ts';

export const diffPluginKey = new PluginKey('diffDecorations');

export interface DiffRange {
  reviewId: string;
  paragraphId: string;
  diffs: DiffItem[];
  from: number;
  to: number;
}

interface DiffPluginState {
  ranges: Map<string, DiffRange>;
  decorationSet: DecorationSet;
}

function rebuildDecorationSet(doc: any, ranges: Map<string, DiffRange>): DecorationSet {
  const allDecorations: Decoration[] = [];

  for (const range of ranges.values()) {
    let pos = range.from;
    for (const diff of range.diffs) {
      if (diff.type === 'equal') {
        pos += diff.text.length;
      } else if (diff.type === 'delete' && !diff.rejected) {
        decorationsPush(allDecorations, pos, diff, range.reviewId);
        pos += diff.text.length;
      } else if (diff.type === 'delete' && diff.rejected) {
        pos += diff.text.length;
      } else if (diff.type === 'insert' && !diff.rejected) {
        allDecorations.push(
          Decoration.widget(pos, () => {
            const span = document.createElement('span');
            span.className = 'diff-insert-widget';
            span.textContent = diff.text;
            span.dataset.reviewId = range.reviewId;
            return span;
          })
        );
      }
    }
  }

  return DecorationSet.create(doc, allDecorations);
}

function decorationsPush(
  decorations: Decoration[],
  pos: number,
  diff: DiffItem,
  reviewId: string
) {
  decorations.push(
    Decoration.inline(pos, pos + diff.text.length, {
      class: 'diff-delete',
      'data-review-id': reviewId,
    })
  );
}

export function createDiffPlugin() {
  return new Plugin({
    key: diffPluginKey,
    state: {
      init(): DiffPluginState {
        return { ranges: new Map(), decorationSet: DecorationSet.empty };
      },
      apply(tr, pluginState): DiffPluginState {
        const meta = tr.getMeta(diffPluginKey);
        if (!meta) {
          if (tr.docChanged) {
            const mappedRanges = new Map<string, DiffRange>();
            for (const [id, range] of pluginState.ranges) {
              const from = tr.mapping.map(range.from);
              mappedRanges.set(id, { ...range, from });
            }
            return {
              ranges: mappedRanges,
              decorationSet: rebuildDecorationSet(tr.doc, mappedRanges),
            };
          }
          return pluginState;
        }

        if (meta.clearAll) {
          return { ranges: new Map(), decorationSet: DecorationSet.empty };
        }

        if (meta.clearReviewId) {
          const ranges = new Map(pluginState.ranges);
          ranges.delete(meta.clearReviewId);
          return {
            ranges,
            decorationSet: rebuildDecorationSet(tr.doc, ranges),
          };
        }

        if (meta.addRange) {
          const range = meta.addRange as DiffRange;
          const ranges = new Map(pluginState.ranges);
          ranges.set(range.reviewId, range);
          return {
            ranges,
            decorationSet: rebuildDecorationSet(tr.doc, ranges),
          };
        }

        if (meta.updateRange) {
          const range = meta.updateRange as DiffRange;
          const ranges = new Map(pluginState.ranges);
          ranges.set(range.reviewId, range);
          return {
            ranges,
            decorationSet: rebuildDecorationSet(tr.doc, ranges),
          };
        }

        return pluginState;
      },
    },
    props: {
      decorations(state) {
        return this.getState(state).decorationSet;
      },
    },
  });
}

export function buildDecorations(
  doc: any,
  from: number,
  diffs: DiffItem[]
): DecorationSet {
  const decorations: Decoration[] = [];
  let pos = from;

  for (const diff of diffs) {
    if (diff.type === 'equal') {
      pos += diff.text.length;
    } else if (diff.type === 'delete' && !diff.rejected) {
      decorations.push(
        Decoration.inline(pos, pos + diff.text.length, {
          class: 'diff-delete',
        })
      );
      pos += diff.text.length;
    } else if (diff.type === 'delete' && diff.rejected) {
      pos += diff.text.length;
    } else if (diff.type === 'insert' && !diff.rejected) {
      decorations.push(
        Decoration.widget(pos, () => {
          const span = document.createElement('span');
          span.className = 'diff-insert-widget';
          span.textContent = diff.text;
          return span;
        })
      );
    }
  }

  return DecorationSet.create(doc, decorations);
}

export function applyDiffForReview(
  editor: { view: { state: any; dispatch: (tr: any) => void } },
  reviewId: string,
  paragraphId: string,
  paragraphFrom: number,
  diffs: DiffItem[]
) {
  const { state, dispatch } = editor.view;
  const tr = state.tr.setMeta(diffPluginKey, {
    addRange: {
      reviewId,
      paragraphId,
      diffs,
      from: paragraphFrom,
      to: paragraphFrom,
    } satisfies DiffRange,
  });
  dispatch(tr);
}

export function clearReviewDecorations(
  editor: { view: { state: any; dispatch: (tr: any) => void } },
  reviewId?: string
) {
  const { state, dispatch } = editor.view;
  const meta = reviewId ? { clearReviewId: reviewId } : { clearAll: true };
  dispatch(state.tr.setMeta(diffPluginKey, meta));
}

export function updateReviewDecorations(
  editor: { view: { state: any; dispatch: (tr: any) => void } },
  range: DiffRange
) {
  const { state, dispatch } = editor.view;
  dispatch(state.tr.setMeta(diffPluginKey, { updateRange: range }));
}
