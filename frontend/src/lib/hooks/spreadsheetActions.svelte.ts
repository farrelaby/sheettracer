import {
  SheetsService,
  ScanService,
} from "../../../bindings/sheettracer/internal/services";
import { TriggeredBy } from "../../../bindings/sheettracer/internal/db/models";
import { spreadsheetStore } from "../spreadsheets.svelte";
import { scanRefresh } from "../scanRefresh.svelte";
import { toast } from "../toast.svelte";

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
    try {
      toast.info(`Scanning ${title}...`);
      const result = await ScanService.ScanSpreadsheet(
        id,
        TriggeredBy.TriggerManual,
      );
      // The scan stamps last_scan_at server-side; reload the store so the
      // sidebar/graph pick it up, and bump the refresh revision so an open
      // inspector refetches without requiring reselect.
      spreadsheetStore.set((await SheetsService.List()) ?? []);
      scanRefresh.bump();
      toast.success(`Scanned ${title} \u2014 ${result?.tabsScanned ?? 0} tabs`);
    } catch (e) {
      console.error("Scan failed:", e);
      toast.error(`Scan failed: ${e}`);
    }
  }

  return { addSheet, removeSheet, scanSpreadsheet };
}
