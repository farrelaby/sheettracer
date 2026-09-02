<script lang="ts">
  import VisibilityIcon from "./VisibilityIcon.svelte";
  import { selectionStore, type Selection } from "@/lib/selection.svelte";

  import {
    SheetsService,
    type DbMetadata as Metadata,
  } from "../../bindings/sheettracer/internal/services";

  import { Browser } from "@wailsio/runtime";

  const currentSelection = $derived(selectionStore.value);

  let node = $derived(
    selectionStore.value?.type === "node"
      ? await SheetsService.FetchDbMetadata(Number(selectionStore.value.id))
      : null,
  );

  const visibilityLabel: Record<string, string> = {
    public: "Public",
    "link-only": "Link only",
    private: "Private",
    unknown: "Unknown",
  };

  function formatDateTime(iso: string): string {
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

  function onSelect(s: Selection) {
    selectionStore.set(s);
  }

  function onClose() {
    selectionStore.set(null);
  }

  $effect(() => {
    console.log(currentSelection);

    console.log(node);
  });
</script>

<aside
  class="absolute inset-y-0 right-0 z-20 flex w-80 flex-col border-l border-slate-800 bg-slate-900/85 backdrop-blur"
>
  <header class="flex items-center justify-between px-4 pb-2 pt-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
      {currentSelection?.type === "node" ? "Spreadsheet" : "Dependency"}
    </p>
    <button
      onclick={onClose}
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
    <!-- {#if loading}
      <div class="flex items-center justify-center py-8">
        <p class="text-xs text-slate-500">Loading...</p>
      </div> -->
    {#if node}
      <!-- ERROR: node title, etc wont refresh when id changed -->
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
