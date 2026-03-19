import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet, type EditorView } from '@tiptap/pm/view';
import { isReady, testSpelling, tokenize } from './spellChecker';

export const spellPluginKey = new PluginKey('spellcheck');

const DEBOUNCE_MS = 400;

function buildSpellDecorations(doc: any): DecorationSet {
  if (!isReady()) return DecorationSet.empty;

  const decorations: Decoration[] = [];

  doc.descendants((node: any, pos: number) => {
    if (!node.isTextblock) return;

    const text = node.textContent;
    if (!text) return;

    const blockStart = pos + 1;
    const tokens = tokenize(text);

    for (const token of tokens) {
      if (!testSpelling(token.word)) {
        decorations.push(
          Decoration.inline(blockStart + token.from, blockStart + token.to, {
            class: 'spell-error',
            nodeName: 'span',
          })
        );
      }
    }
  });

  return DecorationSet.create(doc, decorations);
}

export function createSpellcheckPlugin() {
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;

  return new Plugin({
    key: spellPluginKey,
    state: {
      init(_, state) {
        return buildSpellDecorations(state.doc);
      },
      apply(tr, decorationSet, _oldState, newState) {
        if (tr.getMeta(spellPluginKey)?.forceUpdate) {
          return buildSpellDecorations(newState.doc);
        }
        if (!tr.docChanged) {
          return decorationSet;
        }
        return decorationSet.map(tr.mapping, tr.doc);
      },
    },
    view(editorView: EditorView) {
      return {
        update(view: EditorView, prevState) {
          if (!view.state.doc.eq(prevState.doc)) {
            if (debounceTimer) clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
              view.dispatch(
                view.state.tr.setMeta(spellPluginKey, { forceUpdate: true })
              );
            }, DEBOUNCE_MS);
          }
        },
        destroy() {
          if (debounceTimer) clearTimeout(debounceTimer);
        },
      };
    },
    props: {
      decorations(state) {
        return this.getState(state);
      },
    },
  });
}
