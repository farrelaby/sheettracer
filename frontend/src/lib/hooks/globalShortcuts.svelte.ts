import { onMount } from "svelte";
import { selectionStore } from "../selection.svelte";

export function useGlobalShortcuts() {
  onMount(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") selectionStore.set(null);
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });
}
