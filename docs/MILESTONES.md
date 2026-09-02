# SheetTracer Milestones

Implementation roadmap. Each milestone has deliverables and acceptance criteria. Commits should be small and focused (one logical change per commit) per the project's git convention.

> Milestone 0 status: **done** — `go.mod` is `module sheettracer`, bindings live under `frontend/bindings/sheettracer`, template demo code is removed.

## Milestone 0 — Scaffold cleanup

**Deliverables**
- `go.mod` module renamed to `sheettracer`; bindings regenerate under `frontend/src/bindings/sheettracer`.
- App metadata in `main.go`: `Name: "SheetTracer"`, `Description`, window title/size.
- Remove template `GreetService` demo event/time ticker.
- `wails3 dev` runs the empty shell.

**Acceptance**: `wails3 dev` opens a SheetTracer window; template demo code is gone; `go build ./...` passes.

## Milestone 1 — Google OAuth

**Deliverables**
- Google Cloud project + consent screen + Desktop OAuth client (see `docs/GOOGLE_OAUTH.md`).
- `internal/oauth/`: loopback server, PKCE code exchange, token refresh, keyring-backed token persistence.
- `internal/db`: bootstrap — connection (WAL + pragmas), `user_version` migrations, `settings` KV repo.
- `internal/keyring`: OS-keyring token store (`oauth.refresh_token` / `oauth.token`) with file-backend fallback.
- `OAuthService` with `Connect()`, `Status()`, `Disconnect()`.
- Frontend `OAuthStatus.svelte` with all four states.
- `auth:state` event.

**Acceptance**: first run → "Connect Google" → browser consent → back in app showing account email; relaunch stays connected; refresh works.

## Milestone 2 — Add spreadsheet + metadata

**Deliverables**
- `internal/sheets/`: parse Google Sheets URL → id (strip query/hash params); `spreadsheets.get` metadata (title, tabs) + `drive.files.get` (`version`, `modifiedTime`, `shared`, `capabilities`).
- `internal/db/`: `SpreadsheetRepo` (UpsertByGoogleID, GetByGoogleID, ListTracked, Delete) + `TabRepo` (UpsertAll, ListBySpreadsheet) — sqlx-backed repositories.
- `SheetsService.Add(url, notes)`: `files.get` first — on error (404 `notFound`) reject with "no access / not found"; on success upsert metadata + `visibility` (via `shared` + `capabilities`), dedup on `google_id`.
- `SidebarPanel.svelte`: calls `SheetsService.List()` on mount, `Add()` with 404 error state, `Remove()` to delete.

**Acceptance**: paste a link → workbook appears in the list with title/tabs + visibility icon; a link to a sheet you can't access shows "no access / not found"; adding the same link twice dedups.

## Milestone 3 — Parallel scanner + graph storage

**Deliverables**
- `internal/scan/`: IMPORTRANGE extractor, chunked `values.batchGet` (FORMULA render, field-masked), worker pool + results channel + single merge writer.
- Edge delete/insert lifecycle; `scan_cache` writes; `scan_runs` logging.
- Preflight refreshes `visibility` (same `files.get`); merge resolves external target titles/visibility (cached `files.get`, 404 → `unknown`).
- `ScanService.Rescan()` and `RescanOne(id)`; `scan:progress` + `graph:updated` events.
- Per-workbook 403 handling → `status='partial'`; `GraphPayload` nodes carry `visibility`.

**Acceptance**: add 2–3 spreadsheets with cross-references → edges persisted + target leaves show title/visibility; a removed IMPORTRANGE disappears after rescan; a no-access sheet fails gracefully; a target you can't access shows `unknown`.

## Milestone 4 — Graph view

**Deliverables**
- Add Cytoscape.js + `fcose`.
- `GraphCanvas.svelte`: instance-in-`onMount`, `cy.batch` updates, fan-in node sizing via `mapData`, cluster coloring, external-vs-internal edges, hover labels, focus-on-hover.
- `NodeDetailPanel.svelte`: imports, fan-in, visibility badge, modified time, open-in-Google, leaf → "Track this tab".
- `GraphLegend.svelte` (incl. visibility icons); visibility icon overlays on nodes.
- `GraphService.Graph()` returns `GraphPayload` with `visibility` per node (`docs/FRONTEND.md`).

- **View menu**: Toggle Sidebar (`CmdOrCtrl+B`), Toggle Inspector (`CmdOrCtrl+Shift+I`), separator, Zoom In (`CmdOrCtrl+=`), Zoom Out (`CmdOrCtrl+-`), Reset Zoom (`CmdOrCtrl+0`). Menu items emit Wails events; `App.svelte` and `GraphCanvas.svelte` listen and react.
- **Help menu**: Documentation (opens `docs/` URL or local path via `app.Browser.OpenURL`), Report Issue (opens GitHub issues page).

**Acceptance**: graph renders Obsidian-style; node size reflects inbound dependencies; every node shows its visibility icon; click shows detail (incl. `unknown`/no-access state); leaf offers "Track this tab"; layout re-runs on update; View menu toggles sidebar/inspector and controls zoom; keyboard shortcuts work; Help menu links open in browser.

## Milestone 5 — Cache revalidation + scan triggers

**Deliverables**
- Preflight via `drive.files.get` `version` comparison; skip unchanged workbooks (cache hits).
- `scan_cache` read/write + fingerprint invalidation.
- Triggers: manual rescan, rescan-on-launch (background), rescan-on-add.
- `tabs_skipped` counter surfaced in UI/scan runs.

**Acceptance**: launch scan with nothing changed → near-zero cell reads (all skipped); touch one workbook → only it re-reads.

## Milestone 6 — Polish + release

**Deliverables**
- CSP locked down, no `{@html}`, error/empty states, loading states.
- `docs/CI.md` workflow live: tag → matrix build → GitHub release.
- README/CHANGELOG up to date; `wails3 build` produces release binaries.

**Acceptance**: a tagged release produces Linux/macOS/Windows artifacts; app passes a fresh-install smoke test.

## Backlog (post-MVP)

- More import types (QUERY, IMPORTXML) via the `Extractor` interface.
- Tab-level drill-down toggle in the graph.
- App verification for the "unverified app" warning + >100 users.
- Headless server mode (`build:server`) as a shared/team instance.
- macOS/Windows code signing + notarization.
