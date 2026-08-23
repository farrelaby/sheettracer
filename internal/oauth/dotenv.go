//go:build !production

package oauth

import "github.com/joho/godotenv"

// LoadDotEnv reads a gitignored .env from the working directory so dev builds
// pick up SHEETTRACER_GOOGLE_* without shell exports. It never overrides
// variables already set in the environment and no-ops when the file is
// absent. Production builds (dotenv_prod.go) replace this with a no-op so a
// stray local .env can never redirect release credentials.
func LoadDotEnv() {
	_ = godotenv.Load()
}
