# SheetTracer Google OAuth

SheetTracer reads spreadsheet contents (formulas, for IMPORTRANGE detection) and file metadata (change detection). This requires OAuth — **there is no anonymous API path for reading formula cells**, even for public spreadsheets. The CSV export endpoint has no auth but returns values only (never formulas), so it cannot power this app.

## Scopes

| Scope | Purpose | Sensitivity |
|---|---|---|
| `https://www.googleapis.com/auth/spreadsheets.readonly` | Read cell data + formulas | non-sensitive (read-only) |
| `https://www.googleapis.com/auth/drive.metadata.readonly` | Read `version`/`modifiedTime` for change detection | non-sensitive (metadata only) |

Both are least-privilege. We never request write access.

## Author setup (one-time, Google Cloud Console)

1. Create a project at <https://console.cloud.google.com>.
2. Enable the **Google Sheets API** and **Google Drive API**.
3. **OAuth consent screen**:
   - App name: SheetTracer, logo, support email, developer contact.
   - User type: **External** (for distribution). For personal/early use, External in **Testing** mode with your account on the test-user allowlist.
   - Scopes: the two above.
4. **Credentials → Create credentials → OAuth client ID** → type **Desktop app**.
5. Download `client_secret.json`; embed the client ID/secret into the app (see below).

## Desktop client oddity

For the **Desktop app** client type the `client_secret` is not a true secret — it ships inside the distributed binary. Google treats installed apps this way by design. Consequences:

- Don't try to hide the secret; do keep **scopes read-only** so the blast radius of exposure is minimal.
- Never let user *tokens* leak — those are per-user and sensitive, stored locally in SQLite (`settings` table).

## Auth flow (loopback)

Desktop apps have no hosted callback, so we use the **loopback redirect** pattern:

```
1. Backend starts a local HTTP listener on 127.0.0.1:<port>
2. Opens the default browser to Google's consent URL
   (redirect_uri = http://127.0.0.1:<port>)
3. User signs in, approves scopes
4. Google redirects the browser to the loopback URI with ?code=...
5. Backend captures the code, exchanges it for tokens, closes the listener
6. Refresh token persists in SQLite settings
```

Implementation notes:
- Use a random loopback port per attempt; parse `code` (and `error`) from the redirected URL.
- Use **PKCE** for the code exchange.
- Wails gives you `application.BrowserOpenURL()` (or the runtime equivalent) to open the consent page.

## Token handling

- **Access token**: short-lived (~1h), refreshed automatically by the OAuth client from the refresh token.
- **Refresh token**: stored in `settings` under `oauth.refresh_token`. It is the long-lived credential.
- Persist as raw JSON token payload so reserialization round-trips cleanly.
- **On expiry/invalid grant**: emit `auth:state` with `connected=false, status=expired`; frontend shows the reconnect view.

## App states (frontend)

| State | Meaning | UI |
|---|---|---|
| `disconnected` | No token stored | "Connect Google" button |
| `connecting` | Loopback flow in progress | Spinner + "complete in browser" |
| `connected` | Valid token | Account email, "Reconnect / Disconnect" |
| `expired` | Refresh token rejected | "Reconnect" prompt |

## Verification / publishing (when scaling beyond early users)

While an app is **unverified**:

- Users see the **"Google hasn't verified this app"** interstitial — the main source of auth anxiety.
- There is a **100-new-user cap**.
- In **Testing** mode only allowlisted test users can authorize.

To remove the warning screen and cap, apply for **verification** (requires: domain ownership, privacy policy URL, terms; some sensitive/restricted scopes need review). `spreadsheets.readonly` and `drive.metadata.readonly` are non-sensitive, which keeps verification light. For internal-only use, an **Internal** project skips verification entirely.

## Product UX guidance

- The connect button should read **"Connect Google"** — copy users already trust — with a one-line note: "Read-only access. SheetTracer never modifies your sheets."
- The OAuth dance happens once per machine; subsequent launches are silent.
