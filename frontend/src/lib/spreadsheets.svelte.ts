import type cytoscape from "cytoscape";
import { type Spreadsheet } from "../../bindings/sheettracer/internal/db";

class SpreadsheetState {
  value = $state<Spreadsheet[]>([]);
  loading = $state(false);

  set(value: Spreadsheet[]) {
    this.value = value;
  }

  isEmpty() {
    return this.value.length === 0;
  }

  remove(spreadsheetId: number) {
    this.value = this.value.filter((s) => s.ID !== spreadsheetId);
  }

  append(spreadsheet: Spreadsheet) {
    this.value = [spreadsheet, ...this.value];
  }

  get graphElements(): cytoscape.ElementDefinition[] {
    return this.value.map((s) => ({
      data: {
        id: s.ID.toString(),
        label: s.Title,
      },
      classes: `${s.IsTracked ? "tracked" : "external"} ${s.Visibility}`,
    }));
  }

  setLoading(loading: boolean) {
    this.loading = loading;
  }
}

export const spreadsheetStore = new SpreadsheetState();
