export interface VersionEntry {
  id: string;
  timestamp: number;
  paragraphId: string;
  original: string;
  improved: string;
  summary: string;
  mode?: string;
}

class VersionHistoryStore {
  entries: VersionEntry[] = $state([]);

  addEntry(entry: Omit<VersionEntry, 'id' | 'timestamp'>) {
    const id = `ver-${Date.now()}`;
    this.entries.unshift({ ...entry, id, timestamp: Date.now() });
    if (this.entries.length > 100) {
      this.entries = this.entries.slice(0, 100);
    }
  }

  clear() {
    this.entries = [];
  }
}

export const versionHistoryStore = new VersionHistoryStore();
