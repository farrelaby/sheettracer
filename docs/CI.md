# SheetTracer CI/CD

**Design decision: no local cross-compilation, no Docker/podman.** Build on native GitHub Actions runners per platform. This is the officially recommended path (avoids the ~800MB `wails-cross` Docker image, avoids macOS SDK licensing handling, and produces properly code-signed binaries).

Local dev: `wails3 build` targets Linux natively; `GOOS=windows wails3 build` cross-compiles Windows from Linux for free (no CGO needed for Windows). macOS is always produced on a `macos-latest` runner.

## Workflow

`.github/workflows/release.yml` — a tag-triggered release build.

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
          - os: macos-latest
            goos: darwin
          - os: windows-latest
            goos: windows

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install Wails CLI
        run: go install github.com/wailsapp/wails/v3/cmd/wails3@latest
      - name: Install Task
        uses: arduino/setup-task@v2
      - name: Build
        run: wails3 build
      - uses: actions/upload-artifact@v4
        with:
          name: sheettracer-${{ matrix.goos }}
          path: bin/

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          path: artifacts
          merge-multiple: true
      - uses: softprops/action-gh-release@v2
        with:
          files: artifacts/**
          generate_release_notes: true
```

## Code signing

- **Windows**: optionally sign via `signtool` / an EV or OV cert stored as a GitHub secret.
- **macOS**: sign + notarize on the `macos-latest` runner (has Xcode + `codesign`). Requires Apple Developer certificates in GitHub secrets (`P12_BASE64`, `NOTARIZATION_*`).
- **Linux**: unsigned `.AppImage`/tarball is fine for early releases.

Signing is a follow-up; the base workflow above works unsigned for the first release.

## Notes

- `wails3 build` on each runner produces a native binary; no `setup:docker`, no `wails-cross` image, no podman/docker alias hacks.
- `frontend/dist` is built by `wails3 build` (the Taskfile handles the frontend build), so no separate frontend job is needed.
- A plain **push** (non-tag) workflow can reuse the same matrix for CI smoke builds (`wails3 dev`/`wails3 build` as a compile check).
