export type ToastKind = "success" | "error" | "info";

export type ToastPosition =
  | "bottom-right"
  | "bottom-left"
  | "top-right"
  | "top-left"
  | "bottom-center"
  | "top-center";

export interface Toast {
  id: number;
  kind: ToastKind;
  message: string;
  duration: number;
}

class ToastState {
  toasts = $state<Toast[]>([]);
  position = $state<ToastPosition>("bottom-right");
  private nextId = 0;

  setPosition(pos: ToastPosition) {
    this.position = pos;
  }

  add(kind: ToastKind, message: string, duration = 3000) {
    const id = this.nextId++;
    this.toasts = [...this.toasts, { id, kind, message, duration }];
    if (duration > 0) {
      setTimeout(() => this.dismiss(id), duration);
    }
  }

  success(message: string, duration?: number) {
    this.add("success", message, duration);
  }

  error(message: string, duration?: number) {
    this.add("error", message, duration);
  }

  info(message: string, duration?: number) {
    this.add("info", message, duration);
  }

  dismiss(id: number) {
    this.toasts = this.toasts.filter((t) => t.id !== id);
  }
}

export const toast = new ToastState();
