<script lang="ts">
  import VisibilityIcon from "./VisibilityIcon.svelte";
  import { getNodeDetail, getEdgeDetail, visibilityLabel } from "./demo-data";
  import { selection, type Selection } from "@/lib/selection.svelte";

  const currentSelection = $derived(selection.value!);
  const node = $derived(
    currentSelection?.type === "node"
      ? getNodeDetail(currentSelection.id)
      : undefined,
  );
  const edge = $derived(
    currentSelection?.type === "edge"
      ? getEdgeDetail(currentSelection.id)
      : undefined,
  );

  function onSelect(s: Selection) {
    selection.set(s);
  }

  function onClose() {
    selection.set(null);
  }
</script>

<aside
  class="absolute inset-y-0 right-0 z-20 flex w-80 flex-col border-l border-slate-800 bg-slate-900/85 backdrop-blur"
>
  <header class="flex items-center justify-between px-4 pb-2 pt-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
      {currentSelection?.type === "node"
        ? node?.kind === "external"
          ? "External sheet"
          : "Sheet"
        : "Dependency"}
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
    {#if node}
      <h2 class="text-base font-semibold text-slate-100">{node.title}</h2>
      <div class="mt-1 flex items-center gap-1.5 text-xs text-slate-400">
        <VisibilityIcon visibility={node.visibility} />
        {visibilityLabel[node.visibility]}
      </div>

      {#if node.kind === "external"}
        <p
          class="mt-4 rounded-md border border-slate-800 bg-slate-950/50 p-3 text-xs leading-relaxed text-slate-500"
        >
          This sheet is referenced by an IMPORTRANGE but isn't tracked yet.
          Track it to scan its own imports.
        </p>
        <button
          onclick={() => undefined}
          class="mt-3 w-full rounded-md bg-purple-600 px-3 py-2 text-sm font-medium text-white hover:bg-purple-500"
          >Track this sheet</button
        >
        <button
          class="mt-2 w-full rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-300 hover:bg-slate-800"
          >Open in Google Sheets</button
        >
      {:else}
        <div class="mt-4 grid grid-cols-2 gap-2 text-center">
          <div class="rounded-md border border-slate-800 bg-slate-950/50 p-2.5">
            <p class="text-lg font-semibold text-slate-100">{node.fanIn}</p>
            <p class="text-[11px] text-slate-500">imported from</p>
          </div>
          <div class="rounded-md border border-slate-800 bg-slate-950/50 p-2.5">
            <p class="text-lg font-semibold text-slate-100">{node.fanOut}</p>
            <p class="text-[11px] text-slate-500">imports into</p>
          </div>
        </div>

        <dl class="mt-4 space-y-1 text-xs">
          <div class="flex justify-between">
            <dt class="text-slate-500">Modified</dt>
            <dd class="text-slate-300">{node.modified}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-slate-500">Last scanned</dt>
            <dd class="text-slate-300">{node.lastScanned}</dd>
          </div>
        </dl>

        <div class="mt-4">
          <p class="mb-1.5 text-xs font-medium text-slate-400">Tabs</p>
          <div class="flex flex-wrap gap-1">
            {#each node.tabs as tab (tab)}
              <span
                class="rounded bg-slate-800 px-1.5 py-0.5 text-[11px] text-slate-300"
                >{tab}</span
              >
            {/each}
          </div>
        </div>

        <div class="mt-4">
          <p class="mb-1.5 text-xs font-medium text-slate-400">
            Imports from ({node.importsFrom.length})
          </p>
          {#if node.importsFrom.length === 0}
            <p class="text-xs text-slate-600">
              No sheets import into this one.
            </p>
          {:else}
            <ul class="space-y-1">
              {#each node.importsFrom as dep (dep.edgeId)}
                <li>
                  <button
                    onclick={() => onSelect({ type: "node", id: dep.nodeId })}
                    class="w-full rounded-md border border-transparent px-2 py-1.5 text-left text-xs text-slate-300 hover:border-slate-800 hover:bg-slate-800/40"
                  >
                    {dep.title}
                    <span class="block text-[10px] text-slate-600"
                      >{dep.range}</span
                    >
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>

        <div class="mt-4">
          <p class="mb-1.5 text-xs font-medium text-slate-400">
            Imports into ({node.importsInto.length})
          </p>
          {#if node.importsInto.length === 0}
            <p class="text-xs text-slate-600">
              This sheet doesn't import anything.
            </p>
          {:else}
            <ul class="space-y-1">
              {#each node.importsInto as dep (dep.edgeId)}
                <li>
                  <button
                    onclick={() => onSelect({ type: "node", id: dep.nodeId })}
                    class="w-full rounded-md border border-transparent px-2 py-1.5 text-left text-xs text-slate-300 hover:border-slate-800 hover:bg-slate-800/40"
                  >
                    {dep.title}
                    <span class="block text-[10px] text-slate-600"
                      >{dep.range}</span
                    >
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>

        <div class="mt-5 flex gap-2">
          <button
            class="flex-1 rounded-md border border-slate-700 px-2 py-1.5 text-xs text-slate-300 hover:bg-slate-800"
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
      {/if}
    {:else if edge}
      <div class="flex items-center gap-2 text-sm text-slate-200">
        <button
          onclick={() => onSelect({ type: "node", id: edge.source })}
          class="min-w-0 truncate rounded px-1 py-0.5 text-purple-300 hover:bg-slate-800"
          >{edge.sourceTitle}</button
        >
        <span class="shrink-0 text-slate-500">→</span>
        <button
          onclick={() => onSelect({ type: "node", id: edge.target })}
          class="min-w-0 truncate rounded px-1 py-0.5 text-purple-300 hover:bg-slate-800"
          >{edge.targetTitle}</button
        >
      </div>
      <dl
        class="mt-4 space-y-1.5 rounded-md border border-slate-800 bg-slate-950/50 p-3 text-xs"
      >
        <div class="flex justify-between gap-3">
          <dt class="shrink-0 text-slate-500">Range</dt>
          <dd class="font-mono text-slate-200">{edge.range}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-slate-500">Source tab</dt>
          <dd class="text-slate-300">{edge.sourceTab}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-slate-500">Formulas</dt>
          <dd class="text-slate-300">{edge.formulaCount}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-slate-500">First seen</dt>
          <dd class="text-slate-300">{edge.firstSeen}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-slate-500">Last seen</dt>
          <dd class="text-slate-300">{edge.lastSeen}</dd>
        </div>
      </dl>
      <button
        class="mt-3 w-full rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-300 hover:bg-slate-800"
        >Open target in Google Sheets</button
      >
    {/if}
  </div>
</aside>
