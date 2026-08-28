-- 0001_init.sql
-- Initial SheetTracer schema. See docs/DATABASE.md for the full design and
-- rationale. The OAuth refresh token is NOT stored here — it lives in the OS
-- keyring (internal/keyring); `settings` holds non-secret preferences only.

-- Tracked spreadsheets (graph nodes at file level)
CREATE TABLE spreadsheets (
    id              INTEGER PRIMARY KEY,
    google_id       TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    notes           TEXT NOT NULL DEFAULT '',
    version         TEXT,
    modified_time   TIMESTAMP,
    visibility      TEXT NOT NULL DEFAULT 'private'
                    CHECK (visibility IN ('public', 'link-only', 'private', 'unknown')),
    is_tracked      BOOLEAN NOT NULL DEFAULT 1,
    last_scan_at    TIMESTAMP,
    added_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabs within a workbook (metadata, not graph nodes)
CREATE TABLE sheets (
    id              INTEGER PRIMARY KEY,
    spreadsheet_id  INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    tab_id          INTEGER NOT NULL,
    title           TEXT NOT NULL,
    idx             INTEGER,
    UNIQUE (spreadsheet_id, tab_id)
);

-- Import dependencies (graph edges), one row per IMPORTRANGE formula.
-- Edge width on the graph = COUNT(*) per (source_spreadsheet, target_google_id).
CREATE TABLE edges (
    id                   INTEGER PRIMARY KEY,
    source_spreadsheet   INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    source_tab_id        INTEGER NOT NULL,       -- Google sheetId of the tab containing the formula
    source_cell          TEXT NOT NULL,           -- cell reference, e.g. 'B3'
    target_google_id     TEXT NOT NULL,           -- spreadsheetId parsed from IMPORTRANGE url
    target_range         TEXT,                    -- e.g. 'Sheet1!A1:C10'
    first_seen_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (source_spreadsheet, source_tab_id, target_google_id, target_range)
);
CREATE INDEX idx_edges_source ON edges(source_spreadsheet);
CREATE INDEX idx_edges_target ON edges(target_google_id);
CREATE INDEX idx_edges_pair ON edges(source_spreadsheet, target_google_id);

-- Raw fetched payloads, compressed, so rescans are incremental and memory-light
CREATE TABLE scan_cache (
    id             INTEGER PRIMARY KEY,
    spreadsheet_id INTEGER NOT NULL REFERENCES spreadsheets(id) ON DELETE CASCADE,
    tab_id         INTEGER,
    payload        BLOB NOT NULL,
    fingerprint    TEXT,
    fetched_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (spreadsheet_id, tab_id)
);

-- One row per scan run
CREATE TABLE scan_runs (
    id             INTEGER PRIMARY KEY,
    triggered_by   TEXT NOT NULL CHECK (triggered_by IN ('manual', 'launch', 'add')),
    started_at     TIMESTAMP NOT NULL,
    finished_at    TIMESTAMP,
    status         TEXT NOT NULL CHECK (status IN ('running', 'ok', 'partial', 'error')),
    sheets_scanned INTEGER NOT NULL DEFAULT 0,
    sheets_skipped INTEGER NOT NULL DEFAULT 0,
    error          TEXT
);

-- Key/value store: non-secret preferences and last app state.
-- Secrets (the OAuth refresh token) live in the OS keyring — see "Secrets".
-- Each key corresponds to a settings tab; value is JSONB with that tab's config.
CREATE TABLE settings (
    k TEXT PRIMARY KEY,
    v JSONB NOT NULL DEFAULT '{}'
);
