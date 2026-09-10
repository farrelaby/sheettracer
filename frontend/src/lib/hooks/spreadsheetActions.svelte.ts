import {
  SheetsService,
  ScanService,
} from "../../../bindings/sheettracer/internal/services";
import { TriggeredBy } from "../../../bindings/sheettracer/internal/db/models";
import { spreadsheetStore } from "../spreadsheets.svelte";
import { scanRefresh } from "../scanRefresh.svelte";
import { toast } from "../toast.svelte";
import { SvelteSet } from "svelte/reactivity";

// Shared across all useSpreadsheetActions() instances so the sidebar row
// and the inspector stay in sync about which spreadsheet is scanning.
const scanningIds = new SvelteSet<number>();

function isScanning(id: number): boolean {
  return scanningIds.has(id);
}

// Raw Go errors (e.g. "oauth: token validation: ...", "[about.get]: ...")
// must never reach the toast. Auth/session problems resolve to null
// (silent — OAuthStatus already shows the "Session expired" badge);
// everything else maps to a short friendly message.
function friendlyScanError(e: unknown): string | null {
  const msg = String((e as any)?.message ?? e ?? "").toLowerCase();
  if (
    msg.includes("expired") ||
    msg.includes("refresh") ||
    msg.includes("reauth") ||
    msg.includes("invalid_grant") ||
    msg.includes("token") ||
    msg.includes("oauth") ||
    msg.includes("about.get") ||
    msg.includes("drive service") ||
    msg.includes("401") ||
    msg.includes("403") ||
    msg.includes("unauthorized") ||
    msg.includes("unauthenticated")
  ) {
    return null;
  }
  if (
    msg.includes("no access") ||
    msg.includes("not found") ||
    msg.includes("404")
  ) {
    return "Could not access this sheet — check sharing.";
  }
  if (msg.includes("network") || msg.includes("timeout") || msg.includes("econn")) {
    return "Network error — try again.";
  }
  return "Scan failed — try again.";
}

export function useSpreadsheetActions() {
  async function addSheet(linkInput: string): Promise<boolean> {
    const value = linkInput.trim();
    if (!value) return false;
    if (!value.startsWith("http")) {
      toast.error(
        "Enter a Google Sheets link, e.g. https://docs.google.com/spreadsheets/d/\u2026",
      );
      return false;
    }
    try {
      const sp = await SheetsService.Add(value, "");
      if (sp) {
        spreadsheetStore.append(sp);
        toast.success("Spreadsheet added");
      }
      return true;
    } catch (e: any) {
      const msg = String(e?.message || e);
      if (
        msg.includes("no access") ||
        msg.includes("not found") ||
        msg.includes("404")
      ) {
        toast.error(
          "Could not access this sheet \u2014 no access, or it doesn\u2019t exist.",
        );
      } else {
        toast.error(msg || "Failed to add spreadsheet.");
      }
      return false;
    }
  }

  async function removeSheet(id: number) {
    try {
      await SheetsService.Remove(id);
      spreadsheetStore.remove(id);
      toast.setPosition("top-center");
      toast.success("Spreadsheet removed");
    } catch (e) {
      console.error("Failed to remove:", e);
      toast.error("Failed to remove spreadsheet");
    }
  }

  async function scanSpreadsheet(id: number, title: string) {
    if (scanningIds.has(id)) return;
    scanningIds.add(id);
    try {
      const result = await ScanService.ScanSpreadsheet(
        id,
        TriggeredBy.TriggerManual,
      );
      // The scan stamps last_scan_at server-side; reload the store so the
      // sidebar/graph pick it up, and bump the refresh revision so an open
      // inspector refetches without requiring reselect.
      spreadsheetStore.set((await SheetsService.List()) ?? []);
      scanRefresh.bump();
      toast.success(`Scanned ${title} — ${result?.tabsScanned ?? 0} tabs`);
    } catch (e) {
      console.error("Scan failed:", e);
      const friendly = friendlyScanError(e);
      if (friendly) toast.error(friendly);
    } finally {
      scanningIds.delete(id);
    }
  }

  return { addSheet, removeSheet, scanSpreadsheet, isScanning };
}
