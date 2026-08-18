# SheetTracer Database

SQLite, opened via `modernc.org/sqlite` (pure Go, no CGO). The DB lives in the OS user-data dir (`os.UserConfigDir()/SheetTracer/sheetTracer.db`). WAL mode is enabled for concurrent reads during parallel scans.

## Driver rationale

- `modernc.org/sqlite` is a pure-Go implementation — no CGO, no cross-compile pain on any target (see `docs/CI.md`).
- Alternative `mattn/go-sqlite3` requires CGO; rejected.
- SQLite is proven for exactly this workload (single-process desktop app, graph storage). Precedent: codebase-memory-mcp stores multi-million-node graphs in SQLite.

## Connection settings

- WAL mode (`PRAGMA journal_mode=WAL`)
- `busy_timeout` for single-writer contention
- `foreign_keys = ON`
- One writer connection for merges; read-only connections for parallel reads

## Schema

```sql
-- Tracked spreadsheets (graph nodes at file level)
CREATE TABLE spreadsheets (
    id              INTEGER PRIMARY KEY,
    google_id       TEXT NOT NULL UNIQUE,   -- the spreadsheetId
    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',   -- user-added metadata
    version         TEXT,                   -- Drive files.get version (change signal)
    modified_time   TEXT,                   -- RFC3339, for display
    last_scan_at    TEXT,
    added_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Tabs within a workbook (metadata, not graph nodes)
CREATE TABLE sheets (
    id              INTEGER PRIMARY KEY,
    spreadsheet_id  INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    tab_id          INTEGER,                -- Google sheetId of the tab
    title           TEXT NOT NULL,
    idx             INTEGER,                -- tab order
    UNIQUE (spreadsheet_id, tab_id)
);

-- Import dependencies (graph edges), aggregated per source→target pair
CREATE TABLE imports (
    id                   INTEGER PRIMARY KEY,
    source_spreadsheet   INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    source_tab           TEXT,              -- tab the formula lives in
    target_spreadsheet   TEXT NOT NULL,     -- Google id parsed from the IMPORTRANGE url
    target_range         TEXT,              -- e.g. 'Sheet1!A1:C10'
    formula_count        INTEGER NOT NULL DEFAULT 1,
    first_seen_at        TEXT NOT NULL DEFAULT (datetime('now')),
    last_seen_at         TEXT NOT NULL DEFAULT (datetime('now')),
    seen                 INTEGER NOT NULL DEFAULT 1,  -- 0 = not seen in latest scan
    UNIQUE (source_spreadsheet, target_spreadsheet, target_range, source_tab)
);
CREATE INDEX idx_imports_source ON imports(source_spreadsheet);
CREATE INDEX idx_imports_target ON imports(target_spreadsheet);
CREATE INDEX idx_imports_seen   ON imports(seen);

-- Raw fetched payloads, compressed, so rescans are incremental and memory-light
CREATE TABLE scan_cache (
    id             INTEGER PRIMARY KEY,
    spreadsheet_id INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    tab_id         INTEGER,                -- NULL = workbook-level entry
    payload        BLOB NOT NULL,          -- zstd-compressed JSON
    fingerprint    TEXT,                   -- content hash for invalidation
    fetched_at     TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (spreadsheet_id, tab_id)
);

-- One row per scan run
CREATE TABLE scan_runs (
    id             INTEGER PRIMARY KEY,
    triggered_by   TEXT NOT NULL,          -- 'manual' | 'launch' | 'add'
    started_at     TEXT NOT NULL,
    finished_at    TEXT,
    status         TEXT NOT NULL,          -- 'running' | 'ok' | 'partial' | 'error'
    sheets_scanned INTEGER NOT NULL DEFAULT 0,
    sheets_skipped INTEGER NOT NULL DEFAULT 0,  -- cache hits
    error          TEXT
);

-- Key/value store: OAuth token, preferences, last app state
CREATE TABLE settings (
    k TEXT PRIMARY KEY,
    v TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

## Graph model

- **Node** = a row in `spreadsheets` (file-level granularity; tabs live in `sheets` as metadata).
- **Edge** = a row in `imports`: `source_spreadsheet → target_spreadsheet`.
- **Fan-in** (drives node sizing in the graph UI) is the count of inbound `imports` rows for a spreadsheet — computed at read time:

```sql
SELECT target_spreadsheet AS spreadsheet_id, COUNT(*) AS fan_in
FROM imports
GROUP BY target_spreadsheet;
```

## Edge lifecycle (scan merge)

A scan merge is the only writer during a run:

1. **Upsert** each extracted edge — increment `formula_count`, refresh `last_seen_at`, set `seen=1`.
2. **Mark `seen=0`** on all edges belonging to scanned workbooks before merging.
3. **Purge** edges where `seen=0` after the merge — dependencies that disappeared are removed, so the graph reflects reality.
4. Rewrite `scan_cache` blobs for changed tabs; drop unchanged entries.
5. Update `spreadsheets.version` / `modified_time`.

## Migrations

- Versioned SQL migrations, applied in order at startup.
- Keep them in `internal/db/migrations/`, each file `NNN_description.sql`.
- A `schema_version` row (or SQLite `PRAGMA user_version`) tracks the applied migration.

## Backup

Single `.db` file is trivially copyable. WAL sidecar files (`.db-wal`, `.db-shm`) exist during runtime — back up via a clean checkpoint or `VACUUM INTO` if needed.
