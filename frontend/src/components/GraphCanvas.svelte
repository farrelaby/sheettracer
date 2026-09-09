<script lang="ts">
  import type cytoscape from "cytoscape";
  import { useDataLoad, useGraphMount, useGraphSync } from "@/lib/hooks/graph.svelte";

  let container: HTMLDivElement | undefined = $state();
  let cy = $state<cytoscape.Core | undefined>(undefined);

  useDataLoad();
  useGraphMount(
    () => container,
    (v) => cy = v,
  );
  useGraphSync(() => cy);
</script>

<div
  class="absolute inset-0 cursor-grab"
  style="position:absolute; inset:0;"
  bind:this={container}
></div>

<div
  class="pointer-events-none absolute bottom-3 left-1/2 z-10 flex -translate-x-1/2 items-center gap-3 rounded-full border border-slate-800 bg-slate-900/80 px-4 py-1.5 text-[11px] text-slate-400 backdrop-blur"
>
  <span class="flex items-center gap-1.5"
    ><span class="h-2 w-2 rounded-sm border-2 border-emerald-400 bg-slate-800"
    ></span>editor</span
  >
  <span class="flex items-center gap-1.5"
    ><span class="h-2 w-2 rounded-sm border-2 border-amber-400 bg-slate-800"
    ></span>shared</span
  >
  <span class="flex items-center gap-1.5"
    ><span class="h-2 w-2 rounded-sm border-2 border-red-400 bg-slate-800"
    ></span>private</span
  >
  <span class="flex items-center gap-1.5"
    ><span
      class="h-2 w-2 rounded-sm border-2 border-dashed border-slate-500 bg-slate-950"
    ></span>external</span
  >
</div>
