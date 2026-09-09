import type { Spreadsheet } from "../../bindings/sheettracer/internal/db";

export function formatDateTime(iso: string): string {
  try {
    const d = new Date(iso);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    if (diff < 60_000) return "just now";
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} min ago`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} h ago`;
    return d.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  } catch {
    return iso;
  }
}

export function lastScanText(
  lastScanAt: Spreadsheet["LastScanAt"] | null,
): string {
  if (!lastScanAt) return "never";
  return formatDateTime(lastScanAt);
}

export function statusDot(
  lastScanAt: Spreadsheet["LastScanAt"] | null,
): string {
  return lastScanAt ? "bg-emerald-400" : "bg-slate-500";
}
