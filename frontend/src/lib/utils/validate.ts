export const MIN_WIDTH = 260;
export const MAX_WIDTH = 480;

export function isSheetUrl(value: string): boolean {
  return value.startsWith("http");
}

export function clampWidth(w: number): number {
  return Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, w));
}

export function loadSavedWidth(): number {
  const saved = Number(localStorage.getItem("st.sidebarWidth"));
  return Number.isFinite(saved) ? clampWidth(saved) : 340;
}
