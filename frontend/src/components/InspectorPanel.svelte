<script lang="ts">
  import VisibilityIcon from "./VisibilityIcon.svelte";
  import { selectionStore, type Selection } from "@/lib/selection.svelte";
  import { useInspectorData } from "@/lib/hooks/inspectorData.svelte";
  import { formatDateTime } from "@/lib/utils/format";

  import { Browser } from "@wailsio/runtime";

  const inspector = useInspectorData();
  let node = $derived(inspector.state.node);
  let close = $derived(inspector.close);

  const currentSelection = $derived(selectionStore.value);

  const visibilityLabel: Record<string, string> = {
    public: "Public",
    "link-only": "Link only",
    private: "Private",
    unknown: "Unknown",
  };

  function onSelect(s: Selection) {
    selectionStore.set(s);
  }
</script>

<aside
  class="absolute inset-y-0 right-0 z-20 flex w-64 flex-col border-l border-slate-800 bg-slate-900/85 backdrop-blur"
>
  <header class="flex items-center justify-between px-4 pb-2 pt-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
      {currentSelection?.type === "node" ? "Spreadsheet" : "Dependency"}
    </p>
    <button
      onclick={close}
      title="Close"
      class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-slate-200"
    >
      <svg
        class="h-4 w-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"><path d="M18 6 6 18M6 6l12 12"></path></svg
      >
    </button>
  </header>

  <div class="min-h-0 flex-1 overflow-y-auto px-4 pb-4">
    {#if node}
      <h2 class="text-base font-semibold text-slate-100">{node.Title}</h2>
      <div class="mt-1 flex items-center gap-1.5 text-xs text-slate-400">
        <VisibilityIcon visibility={node.Visibility as any} />
        {visibilityLabel[node.Visibility] ?? node.Visibility}
      </div>

      <dl class="mt-4 space-y-1 text-xs">
        <div class="flex justify-between">
          <dt class="text-slate-500">Modified</dt>
          <dd class="text-slate-300">{formatDateTime(node.ModifiedTime)}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-slate-500">Version</dt>
          <dd class="text-slate-300">{node.Version}</dd>
        </div>
      </dl>

      <div class="mt-4">
        <p class="mb-1.5 text-xs font-medium text-slate-400">Tabs</p>
        {#if node.Tabs && node.Tabs.length > 0}
          <div class="flex flex-wrap gap-1">
            {#each node.Tabs as tab (tab.TabID)}
              <span
                class="rounded bg-slate-800 px-1.5 py-0.5 text-[11px] text-slate-300"
                >{tab.Title}</span
              >
            {/each}
          </div>
        {:else}
          <p class="text-xs text-slate-600">No tabs loaded</p>
        {/if}
      </div>

      <div class="mt-5 flex gap-2">
        <button
          onclick={() =>
            Browser.OpenURL(
              `https://docs.google.com/spreadsheets/d/${node?.SpreadsheetID}`,
            )}
          class="flex-1 cursor-pointer rounded-md border border-slate-700 px-2 py-1.5 text-xs text-slate-300 hover:bg-slate-800"
          >Open in Google</button
        >
        <button
          class="flex-1 rounded-md border border-slate-700 px-2 py-1.5 text-xs text-slate-300 hover:bg-slate-800"
          >Rescan</button
        >
        <button
          class="rounded-md border border-red-900/50 px-2 py-1.5 text-xs text-red-400 hover:bg-red-950/40"
          >Remove</button
        >
      </div>
    {:else}
      <p class="py-8 text-center text-xs text-slate-600">No data available</p>
    {/if}
  </div>
</aside>
