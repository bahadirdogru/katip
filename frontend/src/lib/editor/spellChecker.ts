import { createHunspellFromStrings, type Hunspell } from 'hunspell-wasm';

let hunspell: Hunspell | null = null;
let initPromise: Promise<void> | null = null;

const TURKISH_WORD_RE = /[a-zA-ZçÇğĞıİöÖşŞüÜâÂîÎûÛ]+/g;

export async function initSpellChecker(): Promise<void> {
  if (hunspell) return;
  if (initPromise) return initPromise;

  initPromise = (async () => {
    try {
      const [aff, dic] = await Promise.all([
        fetch('/dictionaries/tr_TR.aff').then((r) => r.text()),
        fetch('/dictionaries/tr_TR.dic').then((r) => r.text()),
      ]);
      hunspell = await createHunspellFromStrings(aff, dic);
    } catch (err) {
      console.error('Yazım denetleyicisi başlatılamadı:', err);
      initPromise = null;
      throw err;
    }
  })();

  return initPromise;
}

export function isReady(): boolean {
  return hunspell !== null;
}

export function testSpelling(word: string): boolean {
  if (!hunspell) return true;
  if (word.length < 2) return true;
  if (/^\d+$/.test(word)) return true;
  return hunspell.testSpelling(word);
}

export function getSuggestions(word: string, max = 5): string[] {
  if (!hunspell) return [];
  try {
    return hunspell.getSpellingSuggestions(word).slice(0, max);
  } catch {
    return [];
  }
}

export function addWord(word: string): void {
  if (!hunspell) return;
  hunspell.addWord(word);
}

export interface SpellToken {
  word: string;
  from: number;
  to: number;
}

export function tokenize(text: string): SpellToken[] {
  const tokens: SpellToken[] = [];
  let match: RegExpExecArray | null;
  TURKISH_WORD_RE.lastIndex = 0;
  while ((match = TURKISH_WORD_RE.exec(text)) !== null) {
    tokens.push({
      word: match[0],
      from: match.index,
      to: match.index + match[0].length,
    });
  }
  return tokens;
}
