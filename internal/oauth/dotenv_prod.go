//go:build production

package oauth

// LoadDotEnv is a deliberate no-op in production builds: release binaries use
// the embedded credentials from credentials_prod.go and must not be
// overridable by a .env file in the working directory.
func LoadDotEnv() {}
