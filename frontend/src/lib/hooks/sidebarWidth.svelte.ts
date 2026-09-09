import { onMount } from "svelte";
import { loadSavedWidth, clampWidth } from "../utils/validate";

export function useSidebarWidth() {
  let width = $state(loadSavedWidth());
  let dragging = $state(false);

  function onHandleDown(e: PointerEvent) {
    dragging = true;
    (e.currentTarget as HTMLDivElement).setPointerCapture(e.pointerId);
  }

  function onHandleMove(e: PointerEvent) {
    if (dragging) width = clampWidth(e.clientX);
  }

  function onHandleUp() {
    if (!dragging) return;
    dragging = false;
    localStorage.setItem("st.sidebarWidth", String(width));
  }

  onMount(() => {
    return () => {
      localStorage.setItem("st.sidebarWidth", String(width));
    };
  });

  return {
    get width() {
      return width;
    },
    set width(v: number) {
      width = clampWidth(v);
    },
    get dragging() {
      return dragging;
    },
    onHandleDown,
    onHandleMove,
    onHandleUp,
  };
}
