import { onMount } from "svelte";
import { selectionStore } from "../selection.svelte";
import { dialogTracker } from "../stores/dialog.svelte";

export function useGlobalShortcuts() {
  onMount(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape" && !dialogTracker.isOpen)
        selectionStore.set(null);
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });
}
