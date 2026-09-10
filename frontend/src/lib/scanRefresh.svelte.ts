// Revision counter that forces data refetches after a scan completes.
//
// The inspector effect only tracks the selection, so reloading
// spreadsheetStore alone would leave an open inspector stale. Bumping this
// revision gives the inspector something to subscribe to. Future background
// triggers (e.g. rescan-on-launch reacting to a `graph:updated` Wails event)
// should call the same bump().
class ScanRefreshState {
  version = $state(0);

  bump() {
    this.version++;
  }
}

export const scanRefresh = new ScanRefreshState();
