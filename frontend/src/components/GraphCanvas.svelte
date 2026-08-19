<script lang="ts">
  import { onMount } from "svelte";
  import cytoscape from "cytoscape";
  import fcose from "cytoscape-fcose";
  import { graphElements } from "./demo-data";
  import { selection } from "@/lib/selection.svelte";

  let container: HTMLDivElement | undefined = $state();
  let cy: cytoscape.Core | undefined;

  onMount(() => {
    if (!container) return;
    cytoscape.use(fcose);
    cy = cytoscape({
      container,
      elements: graphElements,
      style: [
        {
          selector: "node",
          style: {
            "background-color": "#1e293b",
            "border-width": 2,
            "border-color": "#64748b",
            label: "data(label)",
            color: "#e2e8f0",
            "font-size": 11,
            "text-valign": "bottom",
            "text-margin-y": 8,
            "text-wrap": "wrap",
            "text-max-width": 100,
            shape: "round-rectangle",
            width: "mapData(fanIn, 0, 4, 34, 110)",
            height: "mapData(fanIn, 0, 4, 34, 110)",
          },
        },
        { selector: "node.public", style: { "border-color": "#34d399" } },
        { selector: "node.linkOnly", style: { "border-color": "#fbbf24" } },
        { selector: "node.private", style: { "border-color": "#f87171" } },
        { selector: "node.unknown", style: { "border-color": "#64748b" } },
        {
          selector: "node.external",
          style: {
            "background-color": "#0f172a",
            "border-style": "dashed",
            color: "#94a3b8",
            "text-opacity": 0.7,
          },
        },
        {
          selector: "node.ehover",
          style: {
            cursor: "pointer",
            "text-opacity": 0.9,
            opacity: 0.7,
          },
        },
        {
          selector: "node:selected",
          style: { "border-width": 4, "border-color": "#a78bfa" },
        },
        {
          selector: "edge",
          style: {
            width: "mapData(formulaCount, 1, 7, 1, 4)",
            "line-color": "#334155",
            "target-arrow-color": "#334155",
            "target-arrow-shape": "triangle",
            "curve-style": "bezier",
            label: "data(label)",
            color: "#94a3b8",
            "font-size": 9,
            "text-rotation": "autorotate",
            "text-background-color": "#06070f",
            "text-background-opacity": 0.8,
            "text-background-padding": 2,
            "text-opacity": 0,
          },
        },
        {
          selector: "edge.ehover",
          style: { "text-opacity": 1, cursor: "pointer" },
        },
        {
          selector: "edge:selected",
          style: { "line-color": "#a78bfa", "target-arrow-color": "#a78bfa" },
        },
      ] as cytoscape.StylesheetStyle[],
      layout: {
        name: "fcose",
        animate: false,
        idealEdgeLength: 110,
        nodeRepulsion: () => 4800,
        padding: 40,
      } as unknown as cytoscape.LayoutOptions,
      minZoom: 0.2,
      maxZoom: 3,
      wheelSensitivity: 2,
    });

    cy.on("tap", "node,edge", (evt) => {
      const el = evt.target as cytoscape.SingularElementReturnValue;
      // onSelect({ type: el.isNode() ? "node" : "edge", id: el.id() });
      selection.set({ type: el.isNode() ? "node" : "edge", id: el.id() });
    });

    cy.on("tap", (evt) => {
      if ((evt.target as unknown) === cy) selection.set(null);
    });
    cy.on("mouseover", "node", (evt) => evt.target.addClass("ehover"));
    cy.on("mouseout", "node", (evt) => evt.target.removeClass("ehover"));

    cy.on("mouseover", "edge", (evt) => evt.target.addClass("ehover"));
    cy.on("mouseout", "edge", (evt) => evt.target.removeClass("ehover"));
  });
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
    ></span>public</span
  >
  <span class="flex items-center gap-1.5"
    ><span class="h-2 w-2 rounded-sm border-2 border-amber-400 bg-slate-800"
    ></span>link-only</span
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
