import { onMount, untrack } from "svelte";
import {
  SheetsService,
  type DbMetadata as Metadata,
} from "../../../bindings/sheettracer/internal/services";
import { selectionStore } from "../selection.svelte";

export function useInspectorData() {
  // let node = $state<Metadata | null>(null);
  // let loading = $state(false);
  const state = $state<{ node: Metadata | null; loading: boolean }>({
    node: null,
    loading: false,
  });
  const sel = $derived(selectionStore.value);

  $effect(() => {
    if (!sel || sel.type !== "node") {
      state.node = null;
      return;
    }

    let cancelled = false;
    state.loading = true;

    SheetsService.FetchDbMetadata(Number(sel.id))
      .then((result) => {
        if (!cancelled) {
          state.node = result ?? null;
        }
      })
      .catch((e) => {
        console.error("Failed to fetch metadata:", e);
        if (!cancelled) {
          state.node = null;
        }
      })
      .finally(() => {
        if (!cancelled) state.loading = false;
      });

    return () => {
      cancelled = true;
    };
  });

  return {
    state,
    close() {
      selectionStore.set(null);
    },
  };
}
