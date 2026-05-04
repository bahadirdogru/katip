export class SettingsStore {
  fontSize = $state(16);
  fontFamily = $state('ui-sans-serif, system-ui, sans-serif');

  constructor() {
    // Load from localStorage if available
    const savedSize = localStorage.getItem('katip-fontSize');
    if (savedSize) this.fontSize = parseInt(savedSize, 10);

    const savedFont = localStorage.getItem('katip-fontFamily');
    if (savedFont) this.fontFamily = savedFont;
  }

  setFontSize(size: number) {
    if (size < 10) size = 10;
    if (size > 48) size = 48;
    this.fontSize = size;
    localStorage.setItem('katip-fontSize', size.toString());
  }

  setFontFamily(font: string) {
    this.fontFamily = font;
    localStorage.setItem('katip-fontFamily', font);
  }
}

export const settingsStore = new SettingsStore();
