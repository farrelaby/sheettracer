-- 0001_init.sql
-- Initial SheetTracer schema. See docs/DATABASE.md for the full design and
-- rationale. The OAuth refresh token is NOT stored here — it lives in the OS
-- keyring (internal/keyring); `settings` holds non-secret preferences only.

-- Tracked spreadsheets (graph nodes at file level)
CREATE TABLE spreadsheets (
    id              INTEGER PRIMARY KEY,
    google_id       TEXT NOT NULL UNIQUE,   -- the spreadsheetId
    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',   -- user-added metadata
    version         TEXT,                   -- Drive files.get version (change signal)
    modified_time   TEXT,                   -- RFC3339, for display
    visibility      TEXT NOT NULL DEFAULT 'private',  -- 'public'|'link-only'|'private'|'unknown' (Drive permissions)
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

-- Key/value store: non-secret preferences and last app state
CREATE TABLE settings (
    k TEXT PRIMARY KEY,
    v TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);