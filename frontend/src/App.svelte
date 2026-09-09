<script lang="ts">
  import GraphCanvas from "./components/GraphCanvas.svelte";
  import SidebarPanel from "./components/SidebarPanel.svelte";
  import InspectorPanel from "./components/InspectorPanel.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { selectionStore } from "./lib/selection.svelte";
  import { useGlobalShortcuts } from "./lib/hooks/globalShortcuts.svelte";

  let sidebarOpen = $state(true);

  useGlobalShortcuts();
</script>

<div
  class="relative h-screen w-screen select-none overflow-hidden bg-[#06070f] text-slate-200"
>
  <GraphCanvas />

  {#if sidebarOpen}
    <SidebarPanel onClose={() => (sidebarOpen = false)} />
  {:else}
    <button
      onclick={() => (sidebarOpen = true)}
      title="Show sidebar"
      class="absolute left-0 top-1/2 z-30 -translate-y-1/2 rounded-r-md border border-l-0 border-slate-800 bg-slate-900/80 px-1.5 py-3 text-slate-400 backdrop-blur hover:text-slate-200"
    >
      <svg
        class="h-4 w-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"><path d="m9 18 6-6-6-6"></path></svg
      >
    </button>
  {/if}

  {#if !selectionStore.isNull()}
    <InspectorPanel />
  {/if}

  <ToastContainer />
</div>
