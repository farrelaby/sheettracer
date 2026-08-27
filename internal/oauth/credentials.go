//go:build !production

package oauth

// DefaultClientID and DefaultClientSecret are the production Google Cloud
// Desktop OAuth credentials. They are empty here (dev builds) and filled in by
// internal/oauth/credentials_prod.go, which is compiled only with
// `-tags production` — the tag the wails3 release pipeline already passes.
// Dev builds rely on a gitignored .env or SHEETTRACER_GOOGLE_* env vars.
var (
	DefaultClientID     string
	DefaultClientSecret string
)
