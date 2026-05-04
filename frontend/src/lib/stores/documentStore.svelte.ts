export class DocumentStore {
  plotSummary = $state('');

  constructor() {
    const saved = localStorage.getItem('katip-plotSummary');
    if (saved) this.plotSummary = saved;
  }

  setPlotSummary(text: string) {
    this.plotSummary = text;
    localStorage.setItem('katip-plotSummary', text);
  }
}

export const documentStore = new DocumentStore();
