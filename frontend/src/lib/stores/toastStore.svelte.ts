export type ToastType = 'info' | 'success' | 'warning' | 'error';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  duration?: number;
}

class ToastStore {
  toasts: Toast[] = $state([]);

  show(message: string, type: ToastType = 'info', duration = 4000) {
    const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    this.toasts.push({ id, message, type, duration });
    if (duration > 0) {
      setTimeout(() => this.dismiss(id), duration);
    }
    return id;
  }

  success(message: string) {
    return this.show(message, 'success');
  }

  warning(message: string) {
    return this.show(message, 'warning', 5000);
  }

  error(message: string) {
    return this.show(message, 'error', 6000);
  }

  info(message: string) {
    return this.show(message, 'info');
  }

  dismiss(id: string) {
    this.toasts = this.toasts.filter((t) => t.id !== id);
  }
}

export const toastStore = new ToastStore();
