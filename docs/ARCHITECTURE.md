# SheetTracer Architecture

SheetTracer is a Wails v3 desktop app: a Go backend talks to Google's APIs and owns local storage; a Svelte 5 frontend renders the UI and the dependency graph.

## System overview

```
┌────────────────────────────────────────────────────────────┐
│  Frontend (Svelte 5 + Cytoscape.js)                        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │  Connect  │  │  Add      │  │  Graph    │  ┌─────────┐  │
│  │  Google   │  │  Sheets   │  │  View     │  │ Detail  │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  │ Panel   │  │
└────────┼──────────────┼──────────────┼────────└────┬────┘  │
         │ bindings     │ bindings     │ events      │       │
┌────────▼──────────────▼──────────────▼─────────────▼───────┐
│  Go backend (Wails v3)                                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ OAuthSvc │ │ SheetsSvc│ │ ScanSvc  │ │ GraphSvc     │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────┬───────┘   │
│       └────────────┼────────────┼───────────────┘           │
│            ┌───────▼────────────▼───────┐  ┌──────────┐     │
│            │        internal/           │  │  db/     │     │
│            │   oauth · sheets · scan    │  │  SQLite  │     │
│            └────────────────────────────┘  └──────────┘     │
└─────────────────────────────────────────────────────────────┘
         │                                  │
         │ OAuth (loopback)                 │ Sheets/Drive API
         ▼                                  ▼
   user's browser              Google Workspace APIs
```

## Services (Wails v3)

Wails v3 binds Go services to the frontend; generated TypeScript bindings live in `frontend/bindings/`. Services are registered in `main.go` via `application.NewService(...)`.

| Service | Responsibilities |
|---|---|
| `OAuthService` | Drive the connect/disconnect flow, return auth state, expose token for other services |
| `SheetsService` | Add/list/remove spreadsheets, fetch metadata (title, tabs, version, visibility) from Google APIs |
| `ScanService` | Orchestrate scans (parallel worker pool), extract IMPORTRANGE edges, persist results |
| `GraphService` | Query the stored graph for the frontend: nodes, edges, fan-in counts, detail payloads |

## Package layout

Services live in `internal/services/` (Wails v3 bindings). Domain logic lives in `internal/`:

```
internal/
  services/  — Wails-bound services (OAuthService, SheetsService)
  oauth/     — loopback server, token exchange, refresh, keyring persistence
  sheets/    — Google Sheets + Drive API client, URL parser, visibility classifier
  scan/      — IMPORTRANGE extractor, worker pool, scan orchestrator
  graph/     — graph model (nodes/edges), fan-in computation, JSON serialization
  db/        — sqlx connection, migrations, repositories (spreadsheets, tabs, edges,
               scan_cache, scan_runs, settings)
```

## Concurrency model

The scan pipeline is the core concurrency concern. It uses goroutines + channels, with `errgroup` for lifecycle management:

```
scan trigger
      │
      ▼
┌─────────── preflight (parallel) ──────────────┐
│ for each tracked spreadsheet:                 │
│   drive.files.get → version / modifiedTime    │
│   unchanged? → skip (cache hit)               │
└──────────────┬────────────────────────────────┘
               │ dirty workbooks only
               ▼
┌─────────── read phase (worker pool) ──────────┐
│ N workers pull workbook tasks from a channel  │
│ each: spreadsheets.get (metadata)             │
│       values.batchGet (formula cells, chunked)│
│       extract IMPORTRANGE → edge result       │
└──────────────┬────────────────────────────────┘
               │ results channel
               ▼
┌─────────── merge phase (single writer) ───────┐
│ delete old edges · insert formulas · write scan_cache blobs    │
│ recompute fan-in · write scan_cache blobs     │
│ emit graph:updated event                      │
└───────────────────────────────────────────────┘
```

Rules:
- Reads run concurrently; **writes go through a single merge goroutine** (Turso single-writer in WAL mode).
- A scan is idempotent and resumable; each workbook is processed independently so a 403 on one never aborts the whole run.
- Progress events are emitted per-workbook (`scan:progress`) and once at the end (`graph:updated`).

## Data flow: adding a spreadsheet

1. Frontend sends a Google Sheets link to `SheetsService.Add(url, notes)`.
2. `sheets.ParseID` extracts the spreadsheet ID from the URL.
3. `sheets.Client.FetchMetadata` calls `spreadsheets.get` (title, tabs) + `drive.files.get` (version, modifiedTime, shared, capabilities).
4. Workbook is upserted into `spreadsheets` via `SpreadsheetRepo`, tabs via `TabRepo`.
5. The saved `Spreadsheet` is returned to the frontend and prepended to the sidebar list.

## Events (Go → frontend)

Registered with `application.RegisterEvent[T]` and emitted via `app.Event.Emit`:

| Event | Payload | Purpose |
|---|---|---|
| `scan:progress` | `{ spreadsheetId, title, status, tabsScanned }` | Progress during scans |
| `graph:updated` | serialized graph | Frontend refreshes Cytoscape |
| `auth:state` | `{ connected, email, status }` | OAuth status changes |

## Frameworks and key decisions

| Decision | Choice | Rationale |
|---|---|---|
| Desktop shell | Wails v3 | Lean (no Electron), Go backend, stable desktop API in beta |
| Backend language | Go | Concurrency (goroutines/channels) fits parallel sheet reads; user learning Go |
| Frontend | Svelte 5 | Lighter than React, no virtual DOM |
| Graph rendering | Cytoscape.js + fcose | 2D, purpose-built for dependency graphs, fan-in sizing |
| Graph layout | fcose (force-directed, Obsidian-style) | Clustered, node size = connectivity |
| Node sizing | fan-in (inbound degree) | Blast-radius visualization |
| Database | Turso/libSQL (tursogo) | Single-process desktop app; pure-Go via purego, no CGO; MVCC concurrent writes; future cloud sync option |
| DB layer | sqlx + repositories | Struct scanning via `jmoiron/sqlx`; `SpreadsheetRepo`, `TabRepo`, `Settings` follow repo pattern |
| Change detection | Drive `files.get` version/modifiedTime | Workbook-level, documented, reliable |
