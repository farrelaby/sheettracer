# SheetTracer Frontend

Svelte 5 + TypeScript + Vite, generated bindings from `@wailsio/runtime`, Cytoscape.js for the graph.

## Stack

| Piece | Choice | Why |
|---|---|---|
| Framework | Svelte 5 (runes) | Lighter than React, no VDOM |
| Build | Vite 8 | Matches the Wails v3 svelte template |
| Bindings | `@wailsio/runtime` | Generated type-safe Go→JS calls |
| Graph | Cytoscape.js + `fcose` layout | Purpose-built dependency graphs, Obsidian-style force layout |
| Styling | Plain CSS (template default) | Keep it lean; no framework CSS |

## Component tree

```
App.svelte            — shell, routing between views, top-level state
├─ OAuthStatus.svelte — connect/disconnect, auth states (docs/GOOGLE_OAUTH.md)
├─ AddSheet.svelte    — paste link → fetch metadata → notes → add
├─ SpreadsheetList.svelte — tracked workbooks + per-item rescan/remove
├─ ScanProgress.svelte    — live scan:progress events
└─ GraphView.svelte       — the Cytoscape canvas
   ├─ GraphCanvas.svelte  — owns the cytoscape instance
   ├─ NodeDetailPanel.svelte — selected-node details
   └─ GraphLegend.svelte     — color/edge meaning
```

## Cytoscape integration pattern

**Cytoscape is imperative and owns a canvas — do not rebuild it through Svelte reactivity.**

- Create the instance **once** in `onMount` (or `$effect`), keep it in a local/closure variable, not reactive state.
- Push graph changes with `cy.json({ elements })` or `cy.batch(() => ...)`.
- Read user interaction via `cy.on('select', ...)` / `cy.on('tap', ...)` and feed *that* into Svelte state (e.g., selected node id → detail panel).
- Never re-render the canvas on every store change; only mutate the DOM parts that need it.

## Layout & styling

- Layout: **`fcose`** (force-directed with clustering → Obsidian "local graph" look).
- Node size ∝ **fan-in** (inbound import count). Cytoscape style mapping keeps this declarative:

```js
// style mapping from a data field (fanIn) — no JS loop needed
node: {
  'width':  'mapData(fanIn, 0, maxFanIn, 20, 90)',
  'height': 'mapData(fanIn, 0, maxFanIn, 20, 90)',
}
```

- Color by cluster (fcose exposes cluster ids) or by node kind (tracked vs external leaf).
- Edge color: external import vs internal; arrows show direction (`source → target` = source imports target).
- Edge labels (the range, e.g. `Sheet1!A1:C10`) shown **on hover/select only** to avoid clutter.
- Hover → highlight neighborhood, fade the rest (Obsidian focus mode).

## Events

Frontend subscribes via `Events.On` from `@wailsio/runtime`:

| Event | Handler behavior |
|---|---|
| `scan:progress` | Update progress UI (`ScanProgress.svelte`) |
| `graph:updated` | Serialize payload into cytoscape elements, `cy.batch` update, rerun layout |
| `auth:state` | Switch OAuth status view |

## Security

- **Never render sheet contents with `{@html}`.** Cell values are untrusted (an owner can put `<img onerror=...>` in a cell). Render as text only.
- Lock down Wails' CSP so only bundled assets load — no remote content in the webview.
- All Google API calls go through the Go backend; the frontend never holds API credentials.

## Data contract (Go → frontend)

The `graph:updated` payload shape:

```ts
interface GraphPayload {
  nodes: Array<{
    id: string;            // spreadsheet google_id
    label: string;         // title
    fanIn: number;         // inbound import count → node size
    kind: 'tracked' | 'external';
    modifiedTime?: string;
    url?: string;
  }>;
  edges: Array<{
    id: string;
    source: string;        // source google_id
    target: string;        // target google_id
    label: string;         // range, shown on hover
    kind: 'import';        // reserved for future import types
  }>;
}
```

## Open in Google

Node detail panel: "Open in Google Sheets" → `application.BrowserOpenURL(spreadsheet.url)` (via a Go service method; the frontend must not open arbitrary URLs outside the backend's whitelist).
