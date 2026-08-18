# SheetTracer

Track Google Sheets import dependencies — starting with `IMPORTRANGE`. Paste a spreadsheet link, SheetTracer scans the workbook for formulas that pull data from other sheets, stores the dependency graph locally, and renders it as an interactive Obsidian-style graph.

## Stack

| Layer | Choice |
|---|---|
| Desktop shell | Wails v3 (Go backend + OS webview) |
| Backend | Go |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Frontend | Svelte 5 + TypeScript + Vite |
| Graph | Cytoscape.js with `fcose` layout |
| Google access | Sheets API (read-only) + Drive API (metadata only) |

## Prerequisites

- Go 1.24+ (project targets `go 1.25.0`)
- Wails v3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- `wails3 setup` (decline Docker — we build on native runners; see `docs/CI.md`)
- Node.js (used by the Svelte/Vite frontend)
- Linux: `gcc`, `gtk4`, `webkitgtk-6.0` (check with `wails3 doctor`)

## Development

```bash
wails3 dev        # hot-reload frontend + backend
wails3 build      # production build
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — services, data flow, concurrency design
- [Database](docs/DATABASE.md) — SQLite schema and storage decisions
- [Google OAuth](docs/GOOGLE_OAUTH.md) — Cloud Console setup and auth flow
- [Scanning](docs/SCANNING.md) — IMPORTRANGE extraction and cache revalidation
- [Frontend](docs/FRONTEND.md) — Svelte components and the Cytoscape graph
- [CI/CD](docs/CI.md) — release builds on GitHub Actions
- [Milestones](docs/MILESTONES.md) — implementation roadmap
