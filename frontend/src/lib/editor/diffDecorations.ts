import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';
import type { DiffItem } from '../stores/reviewStore';

export const diffPluginKey = new PluginKey('diffDecorations');

export interface DiffRange {
  paragraphId: string;
  reviewId: string;
  diffs: DiffItem[];
  from: number;
  to: number;
}

export function createDiffPlugin() {
  return new Plugin({
    key: diffPluginKey,
    state: {
      init() {
        return DecorationSet.empty;
      },
      apply(tr, set) {
        const meta = tr.getMeta(diffPluginKey);
        if (meta?.clear) {
          return DecorationSet.empty;
        }
        if (meta?.decorations) {
          return meta.decorations;
        }
        return set.map(tr.mapping, tr.doc);
      },
    },
    props: {
      decorations(state) {
        return this.getState(state);
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
    } else if (diff.type === 'delete') {
      decorations.push(
        Decoration.inline(pos, pos + diff.text.length, {
          class: 'diff-delete',
        })
      );
      pos += diff.text.length;
    } else if (diff.type === 'insert') {
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
