# SheetTracer Scanning

The scan pipeline detects `IMPORTRANGE` dependencies across tracked spreadsheets, stores them as edges, and revalidates cheaply via cache. **IMPORTRANGE first; other import types (QUERY, IMPORTXML, etc.) are future extensions.**

## What we extract

`IMPORTRANGE` has the signature `IMPORTRANGE("spreadsheet_url", "range_string")`:

```text
=IMPORTRANGE("https://docs.google.com/spreadsheets/d/ABC123/edit#gid=0", "Sheet1!A1:C10")
```

We only care about the **formula cells** — hence the `fields` masking below. Per formula we extract:

- **source cell** — the cell containing the formula (e.g. `B3`), for "open in Google Sheets" links
- **source tab** — which tab (Google `sheetId`) the formula lives in
- **target spreadsheet id** — parsed from the URL (`/d/<id>/` or `/d/<id>/edit`)
- **target range** — the second argument, e.g. `Sheet1!A1:C10`

Regex/extraction is case-insensitive for `IMPORTRANGE` and tolerant of `'` vs `"` quotes and whitespace.

## Memory-efficient reads

The Google Sheets API returns values as one big JSON payload — there is no streaming endpoint. Memory is controlled by requesting less and discarding fast:

1. **Never request all cells.** `spreadsheets.get` returns metadata; grid data is opt-in.
2. **Field-mask formula cells.** `values.batchGet` with:
   - `valueRenderOption=FORMULA`
   - `majorDimension=ROWS`
   - `fields=valueRanges(values)` (limit response shape)
   Then scan only cells whose string starts with `=`.
3. **Chunk by tab.** Read one tab (or a set of ranges) at a time via `values.batchGet` `ranges`, never the whole workbook.
4. **Scan-and-discard.** Extract the dependency refs, drop the cell data. Only the extracted edges and (optionally) compressed raw blobs are retained.

For very large workbooks the read is chunked by range windows (e.g., batches of rows) to cap per-response memory.

## Parallel reader (goroutines + channels)

See `docs/ARCHITECTURE.md` for the diagram. Summary:

- **Preflight phase**: `drive.files.get` per tracked spreadsheet (parallel) → `version`/`modifiedTime`.
- **Read phase**: worker pool (N workers) pulling workbook tasks off a channel; each worker does its own metadata + formula reads and pushes edge results to a results channel.
- **Merge phase**: a single goroutine consumes the results channel, writes to SQLite, and emits events. Single writer avoids WAL contention.

## Cache revalidation

**Design goal: most rescans touch zero cell data.**

| Signal | Source | Used for |
|---|---|---|
| `version` | `drive.files.get` (Drive API) | Monotonic change detection — the primary skip signal |
| `modifiedTime` | `drive.files.get` | Display ("last modified") + secondary check |
| `permissions` | `drive.files.get` (`fields=permissions(id,type,role)`) | Visibility classification (public/link-only/private) |
| `scan_cache` (SQLite) | local | Reuse of raw payloads when a workbook is clean |

Preflight per workbook:

```
version_changed?  →  no  → skip entirely (cache hit, count as sheets_skipped)
                 →  yes → full metadata + formula re-read
```

Notes:

- **Granularity is per-file, not per-tab or per-cell.** The API exposes no per-tab change timestamps, so we don't chase finer revalidation. This is still correct for us: any formula change necessarily bumps the workbook `version`; value-only changes (e.g., imported data refreshing) do *not* change formulas, so skipping them is correct.
- The Sheets API `spreadsheets.get` resource has **no documented `modifiedTime`/etag** — do not rely on undocumented etag headers. Drive is the documented, reliable source.
- `permissions` comes from the same `files.get` call (extended field mask) — no extra request. Sharing changes don't reliably bump `version`, so visibility is refresh-on-scan, which is fine here.

## Visibility (public vs private)

Derived from `drive.files.get` using `shared` + `capabilities` fields:

| State | Signal | Meaning |
|---|---|---|
| `private` | `shared=false` | Only specific people have access |
| `link-only` | `shared=true, canEdit=false` | Anyone with the link can view |
| `public` | `shared=true, canEdit=true` | Anyone with the link can edit |
| `unknown` | `files.get` fails | See below |

**`unknown` = failed `files.get`.** Drive returns HTTP 404 `notFound` for both "no read access" and "file doesn't exist" — deliberately, so the API doesn't leak whether a file exists. The backend treats both identically; the UI shows a muted "not accessible / not found" state. A genuinely public file is always readable by any authenticated user, so a 404 reliably means private-to-you or nonexistent — safe to classify as `unknown`.

## External target resolution (hybrid)

Tracked sheets' IMPORTRANGE targets that the user hasn't added appear as leaf nodes. During the scan merge:

1. Collect target spreadsheet ids from extracted edges that aren't already in `spreadsheets`.
2. Best-effort `files.get` each (same field mask: `name` + `permissions`) to resolve the **title** and **visibility** for display. Cache the result in `scan_cache` so each new target costs ~1 call, not 1-per-scan.
3. On 404 → store as `unknown` (no title, "not accessible / not found" state).

Rules:

- **Metadata only** — external targets are never content-scanned. "Scanning strangers' sheets is out of scope" (fan-in section below).
- The frontend offers a **"Track this sheet"** action on a leaf node, which runs the normal add flow for that URL — the *only* way a discovered target becomes a scanned node.
- If the user later tracks a target, it upgrades from leaf to a full tracked node; edges to it are unchanged.

## Edge lifecycle (per scan merge)

1. **Delete** all edges of scanned workbooks (clear the slate for those workbooks).
2. **Insert** each extracted formula as a row in `edges` — cell, tab, target, range, fresh `first_seen_at`, `last_seen_at = CURRENT_TIMESTAMP`.
3. Store clean tabs in `scan_cache`; refresh `spreadsheets.version`/`modified_time`; resolve + refresh external target titles/visibility.

## Scan triggers

| Trigger | Behavior |
|---|---|
| **Manual** (user clicks Rescan) | Full preflight on all tracked sheets |
| **On launch** | Same as manual, run in the background; cheap because preflight skips unchanged workbooks |
| **On add** | Preflight + read the newly added workbook (others may also be checked) |

## Fan-in computation

Fan-in = number of inbound edges per target spreadsheet (how many tracked sheets depend on it). Computed from the `edges` table:

```sql
SELECT target_google_id, COUNT(*) AS fan_in
FROM edges GROUP BY target_google_id;
```

This feeds node sizing in the graph UI. Fan-in is computed over **tracked** sheets only — external sheets referenced by IMPORTRANGE appear as leaf nodes with best-effort title + visibility (see "External target resolution" above); they're never scanned unless the user tracks them too. Scanning strangers' sheets is out of scope.

## Error handling

- A 403 (no access) on one workbook marks it failed in the scan run but does **not** abort the run (`scan_runs.status = 'partial'`).
- Transient API errors: retry with backoff per workbook.
- Duplicate URLs: dedup on `google_id` unique constraint.

## Future import types

The extractor interface is designed to be one-per-type:

```text
type Extractor interface {
    Name() string
    // returns source→target edges from a formula string
    Extract(formula string) []Edge
}
```

`ImporrangeExtractor` implements it now; `QueryExtractor`/`ImportxmlExtractor` slot in later without touching the pipeline.
