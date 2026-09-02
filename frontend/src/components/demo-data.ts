import type cytoscape from "cytoscape";

export type Visibility = "private" | "shared" | "editor" | "unknown";

export interface DemoNode {
  id: string;
  title: string;
  kind: "tracked" | "external";
  visibility: Visibility;
  fanIn: number;
  fanOut: number;
  modified: string;
  lastScanned: string;
  tabs: string[];
}

export interface DemoEdge {
  id: string;
  source: string;
  target: string;
  range: string;
  sourceTab: string;
  formulaCount: number;
  firstSeen: string;
  lastSeen: string;
}

export type Selection =
  | { type: "node"; id: string }
  | { type: "edge"; id: string };

export const demoNodes: DemoNode[] = [
  {
    id: "n-fin",
    title: "Finance Dashboard",
    kind: "tracked",
    visibility: "shared",
    fanIn: 4,
    fanOut: 1,
    modified: "2 h ago",
    lastScanned: "10 min ago",
    tabs: ["Overview", "Sheet2", "All"],
  },
  {
    id: "n-sales",
    title: "Sales Tracker",
    kind: "tracked",
    visibility: "private",
    fanIn: 0,
    fanOut: 4,
    modified: "1 d ago",
    lastScanned: "10 min ago",
    tabs: ["Sheet1", "Data", "Leads", "Names"],
  },
  {
    id: "n-mkt",
    title: "Marketing Report",
    kind: "tracked",
    visibility: "private",
    fanIn: 0,
    fanOut: 2,
    modified: "3 d ago",
    lastScanned: "1 h ago",
    tabs: ["Sheet2", "People"],
  },
  {
    id: "n-ops",
    title: "Ops Metrics",
    kind: "tracked",
    visibility: "editor",
    fanIn: 0,
    fanOut: 1,
    modified: "5 d ago",
    lastScanned: "1 h ago",
    tabs: ["All", "Notes"],
  },
  {
    id: "n-prod",
    title: "Product Analytics",
    kind: "tracked",
    visibility: "editor",
    fanIn: 2,
    fanOut: 0,
    modified: "6 h ago",
    lastScanned: "1 h ago",
    tabs: ["Events", "Conversion"],
  },
  {
    id: "n-cs",
    title: "CS Quality",
    kind: "tracked",
    visibility: "shared",
    fanIn: 0,
    fanOut: 1,
    modified: "4 d ago",
    lastScanned: "2 h ago",
    tabs: ["Tickets"],
  },
  {
    id: "n-hr",
    title: "HR Headcount",
    kind: "tracked",
    visibility: "private",
    fanIn: 1,
    fanOut: 1,
    modified: "2 d ago",
    lastScanned: "2 h ago",
    tabs: ["Names"],
  },
  {
    id: "n-ext1",
    title: "Untitled sheet",
    kind: "external",
    visibility: "unknown",
    fanIn: 1,
    fanOut: 0,
    modified: "—",
    lastScanned: "—",
    tabs: [],
  },
  {
    id: "n-ext2",
    title: "Budget Master",
    kind: "external",
    visibility: "shared",
    fanIn: 1,
    fanOut: 0,
    modified: "—",
    lastScanned: "—",
    tabs: [],
  },
  {
    id: "n-ext3",
    title: "Payroll Export",
    kind: "external",
    visibility: "private",
    fanIn: 1,
    fanOut: 0,
    modified: "—",
    lastScanned: "—",
    tabs: [],
  },
  {
    id: "n-ext4",
    title: "CRM Dump",
    kind: "external",
    visibility: "editor",
    fanIn: 1,
    fanOut: 0,
    modified: "—",
    lastScanned: "—",
    tabs: [],
  },
];

export const demoEdges: DemoEdge[] = [
  {
    id: "e-sales-fin",
    source: "n-sales",
    target: "n-fin",
    range: "Sheet1!A1:C10",
    sourceTab: "Sheet1",
    formulaCount: 4,
    firstSeen: "Aug 10",
    lastSeen: "10 min ago",
  },
  {
    id: "e-sales-ext1",
    source: "n-sales",
    target: "n-ext1",
    range: "Data!A1:D20",
    sourceTab: "Data",
    formulaCount: 1,
    firstSeen: "Aug 12",
    lastSeen: "10 min ago",
  },
  {
    id: "e-sales-ext4",
    source: "n-sales",
    target: "n-ext4",
    range: "Leads!A1:B30",
    sourceTab: "Leads",
    formulaCount: 5,
    firstSeen: "Aug 14",
    lastSeen: "10 min ago",
  },
  {
    id: "e-sales-hr",
    source: "n-sales",
    target: "n-hr",
    range: "Names!A1:C15",
    sourceTab: "Names",
    formulaCount: 3,
    firstSeen: "Aug 11",
    lastSeen: "10 min ago",
  },
  {
    id: "e-mkt-fin",
    source: "n-mkt",
    target: "n-fin",
    range: "Sheet2!A1:C10",
    sourceTab: "Sheet2",
    formulaCount: 2,
    firstSeen: "Aug 8",
    lastSeen: "1 h ago",
  },
  {
    id: "e-mkt-ext3",
    source: "n-mkt",
    target: "n-ext3",
    range: "People!A1:B12",
    sourceTab: "People",
    formulaCount: 4,
    firstSeen: "Aug 9",
    lastSeen: "1 h ago",
  },
  {
    id: "e-ops-fin",
    source: "n-ops",
    target: "n-fin",
    range: "All!A1:Z50",
    sourceTab: "All",
    formulaCount: 7,
    firstSeen: "Aug 5",
    lastSeen: "1 h ago",
  },
  {
    id: "e-prod-fin",
    source: "n-prod",
    target: "n-fin",
    range: "Events!A1:D40",
    sourceTab: "Events",
    formulaCount: 6,
    firstSeen: "Aug 6",
    lastSeen: "1 h ago",
  },
  {
    id: "e-cs-prod",
    source: "n-cs",
    target: "n-prod",
    range: "Tickets!A1:E22",
    sourceTab: "Tickets",
    formulaCount: 2,
    firstSeen: "Aug 13",
    lastSeen: "2 h ago",
  },
  {
    id: "e-hr-prod",
    source: "n-hr",
    target: "n-prod",
    range: "Names!A1:B25",
    sourceTab: "Names",
    formulaCount: 1,
    firstSeen: "Aug 7",
    lastSeen: "2 h ago",
  },
  {
    id: "e-fin-ext2",
    source: "n-fin",
    target: "n-ext2",
    range: "Sheet1!A1:B10",
    sourceTab: "Overview",
    formulaCount: 3,
    firstSeen: "Aug 9",
    lastSeen: "10 min ago",
  },
];

const nodeById = new Map(demoNodes.map((n) => [n.id, n]));
const edgeById = new Map(demoEdges.map((e) => [e.id, e]));

export function getNode(id: string): DemoNode | undefined {
  return nodeById.get(id);
}

export function getEdge(id: string): DemoEdge | undefined {
  return edgeById.get(id);
}

export interface NodeDetail extends DemoNode {
  importsFrom: { nodeId: string; edgeId: string; title: string; range: string }[];
  importsInto: { nodeId: string; edgeId: string; title: string; range: string }[];
}

export function getNodeDetail(id: string): NodeDetail | undefined {
  const n = nodeById.get(id);
  if (!n) return undefined;
  const importsFrom = demoEdges
    .filter((e) => e.target === id)
    .map((e) => ({
      nodeId: e.source,
      edgeId: e.id,
      title: nodeById.get(e.source)?.title ?? e.source,
      range: e.range,
    }));
  const importsInto = demoEdges
    .filter((e) => e.source === id)
    .map((e) => ({
      nodeId: e.target,
      edgeId: e.id,
      title: nodeById.get(e.target)?.title ?? e.target,
      range: e.range,
    }));
  return { ...n, importsFrom, importsInto };
}

export interface EdgeDetail extends DemoEdge {
  sourceTitle: string;
  targetTitle: string;
}

export function getEdgeDetail(id: string): EdgeDetail | undefined {
  const e = edgeById.get(id);
  if (!e) return undefined;
  return {
    ...e,
    sourceTitle: nodeById.get(e.source)?.title ?? e.source,
    targetTitle: nodeById.get(e.target)?.title ?? e.target,
  };
}

export const visibilityLabel: Record<Visibility, string> = {
  private: "Private",
  shared: "Shared",
  editor: "Editor",
  unknown: "Unknown",
};

export interface SheetRow {
  id: string;
  title: string;
  visibility: Visibility;
  lastScan: string;
  status: "ok" | "partial" | "error" | "scanning";
}

export const seedSheets: SheetRow[] = demoNodes
  .filter((n) => n.kind === "tracked")
  .map((n) => ({
    id: n.id,
    title: n.title,
    visibility: n.visibility,
    lastScan: n.lastScanned,
    status: "ok",
  }));

export const graphElements: cytoscape.ElementDefinition[] = [
  ...demoNodes.map((n) => ({
    data: { id: n.id, label: n.title, fanIn: n.fanIn },
    classes: `${n.kind} ${n.visibility}`,
  })),
  ...demoEdges.map((e) => ({
    data: {
      id: e.id,
      source: e.source,
      target: e.target,
      label: e.range,
      formulaCount: e.formulaCount,
    },
  })),
];