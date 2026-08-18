# SheetTracer Milestones

Implementation roadmap. Each milestone has deliverables and acceptance criteria. Commits should be small and focused (one logical change per commit) per the project's git convention.

> Setup note: `go.mod` still says `module changeme` from the scaffold — Milestone 0 fixes it to `module sheettracer`.

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
- `internal/oauth/`: loopback server, PKCE code exchange, token refresh, `settings` persistence.
- `OAuthService` with `Connect()`, `Status()`, `Disconnect()`.
- Frontend `OAuthStatus.svelte` with all four states.
- `auth:state` event.

**Acceptance**: first run → "Connect Google" → browser consent → back in app showing account email; relaunch stays connected; refresh works.

## Milestone 2 — Add spreadsheet + metadata

**Deliverables**
- `internal/sheets/`: parse Google Sheets URL → id; `spreadsheets.get` metadata (title, tabs) + `drive.files.get` (`version`, `modifiedTime`).
- `internal/db/`: migrations, connection (WAL), `spreadsheets`/`sheets`/`settings` repositories.
- `SheetsService.Add(url, notes)` upserts and returns metadata; dedup on `google_id`.
- `AddSheet.svelte` + `SpreadsheetList.svelte`.

**Acceptance**: paste a link → workbook appears in the list with title/tabs; adding the same link twice dedups.

## Milestone 3 — Parallel scanner + graph storage

**Deliverables**
- `internal/scan/`: IMPORTRANGE extractor, chunked `values.batchGet` (FORMULA render, field-masked), worker pool + results channel + single merge writer.
- Edge upsert/seen/purge lifecycle; `scan_cache` writes; `scan_runs` logging.
- `ScanService.Rescan()` and `RescanOne(id)`; `scan:progress` + `graph:updated` events.
- Per-workbook 403 handling → `status='partial'`.

**Acceptance**: add 2–3 sheets with cross-references → edges persisted; a removed IMPORTRANGE disappears after rescan; a no-access sheet fails gracefully.

## Milestone 4 — Graph view

**Deliverables**
- Add Cytoscape.js + `fcose`.
- `GraphCanvas.svelte`: instance-in-`onMount`, `cy.batch` updates, fan-in node sizing via `mapData`, cluster coloring, external-vs-internal edges, hover labels, focus-on-hover.
- `NodeDetailPanel.svelte`: imports, fan-in, modified time, open-in-Google.
- `GraphLegend.svelte`.
- `GraphService.Graph()` returns `GraphPayload` (`docs/FRONTEND.md`).

**Acceptance**: graph renders Obsidian-style; node size reflects inbound dependencies; click shows detail; layout re-runs on update.

## Milestone 5 — Cache revalidation + scan triggers

**Deliverables**
- Preflight via `drive.files.get` `version` comparison; skip unchanged workbooks (cache hits).
- `scan_cache` read/write + fingerprint invalidation.
- Triggers: manual rescan, rescan-on-launch (background), rescan-on-add.
- `sheets_skipped` counter surfaced in UI/scan runs.

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
