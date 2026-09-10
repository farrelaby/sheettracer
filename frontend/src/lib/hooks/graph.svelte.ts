import { onMount } from "svelte";
import { SheetsService } from "../../../bindings/sheettracer/internal/services";
import { spreadsheetStore } from "../spreadsheets.svelte";
import { toast } from "../toast.svelte";
import cytoscape from "cytoscape";
import fcose from "cytoscape-fcose";
import { selectionStore } from "../selection.svelte";

export function useDataLoad() {
  onMount(async () => {
    spreadsheetStore.setLoading(true);
    try {
      spreadsheetStore.set((await SheetsService.List()) ?? []);
    } catch (e) {
      console.error("Failed to load spreadsheets:", e);
      toast.error("Failed to load spreadsheets");
    } finally {
      spreadsheetStore.setLoading(false);
    }
  });
}

export function useGraphMount(
  getContainer: () => HTMLDivElement | undefined,
  setCy: (cy: cytoscape.Core) => void,
) {
  onMount(() => {
    const container = getContainer();
    if (!container) return;
    cytoscape.use(fcose);
    const instance = cytoscape({
      container,
      elements: [],
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
        { selector: "node.editor", style: { "border-color": "#34d399" } },
        { selector: "node.shared", style: { "border-color": "#fbbf24" } },
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
          style: {
            "line-color": "#a78bfa",
            "target-arrow-color": "#a78bfa",
          },
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

    instance.on("tap", "node,edge", (evt) => {
      const el = evt.target as cytoscape.SingularElementReturnValue;
      selectionStore.set({
        type: el.isNode() ? "node" : "edge",
        id: el.id(),
      });
    });

    instance.on("tap", (evt) => {
      if ((evt.target as unknown) === instance) selectionStore.set(null);
    });
    instance.on("mouseover", "node", (evt) => evt.target.addClass("ehover"));
    instance.on("mouseout", "node", (evt) => evt.target.removeClass("ehover"));

    instance.on("mouseover", "edge", (evt) => evt.target.addClass("ehover"));
    instance.on("mouseout", "edge", (evt) => evt.target.removeClass("ehover"));

    setCy(instance);

    return () => {
      instance.destroy();
    };
  });
}

export function useGraphSync(
  getCy: () => cytoscape.Core | undefined,
) {
  $effect(() => {
    const cy = getCy();
    if (!cy) return;
    const elements = spreadsheetStore.graphElements;

    const existingIds = new Set(cy.nodes().map((n) => n.id()));
    const nextIds = new Set(elements.map((e) => e.data.id as string));
    const sameStructure =
      existingIds.size === nextIds.size &&
      [...nextIds].every((id) => existingIds.has(id));

    // Same nodes, new data (e.g. after a scan): update in place so dragged
    // positions and viewport survive.
    if (sameStructure) {
      for (const el of elements) {
        const node = cy.getElementById(el.data.id as string);
        node.data(el.data);
        node.removeClass("tracked external private shared editor unknown");
        node.addClass(el.classes as string);
      }
      return;
    }

    cy.elements().remove();
    if (elements.length > 0) {
      cy.add(elements);
      cy.layout({
        name: "fcose",
        animate: false,
        idealEdgeLength: 110,
        nodeRepulsion: () => 4800,
        padding: 40,
      } as unknown as cytoscape.LayoutOptions).run();
    }
  });
}
