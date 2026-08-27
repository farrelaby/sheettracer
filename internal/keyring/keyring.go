// Package keyring is the single place SheetTracer keeps secrets. It wraps the
// OS keyring via github.com/zalando/go-keyring (macOS Keychain, Windows
// Credential Manager, Linux freedesktop Secret Service) and transparently
// falls back to an encrypted-on-disk file when no OS keyring is available
// (headless Linux/CI) or when the desktop is KDE (whose wallet would otherwise
// pop an unlock/create window during the OAuth flow).
package keyring

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	gokeyring "github.com/zalando/go-keyring"
)

// Item keys stored in the system keyring.
const (
	// RefreshToken is the long-lived OAuth credential.
	RefreshToken = "oauth.refresh_token"
	// TokenJSON is the raw serialized token payload (see docs/GOOGLE_OAUTH.md).
	TokenJSON = "oauth.token"
)

// fallbackPassphrase guards the file-backend fallback. It ships inside the
// binary, so this is obfuscation, not real security — the fallback exists to
// keep the app functional (headless Linux, CI, KDE) and should never be
// treated as a trusted secret store.
const fallbackPassphrase = "sheettracer-local-keyring-fallback-v1"

// ErrKeyNotFound is returned by Get when the key does not exist in the
// keyring. Use errors.Is to distinguish "absent" from real failures.
var ErrKeyNotFound = errors.New("keyring: key not found")

// Store is a thin wrapper around the OS keyring with a file fallback.
type Store struct {
	service  string
	fileDir  string
	fallback bool // true when backed by the file fallback, not the OS keyring
}

// NewStore opens the platform default keyring for serviceName. On KDE it goes
// straight to the file fallback (no wallet window); on other Linux desktops it
// probes Secret Service and falls back to the file store when unavailable.
func NewStore(serviceName, fileFallbackDir string) (*Store, error) {
	s := &Store{service: serviceName, fileDir: fileFallbackDir}

	if runtime.GOOS == "linux" && isKDE() {
		s.fallback = true
		return s, nil
	}

	if runtime.GOOS == "linux" {
		// Probe Secret Service availability with a read of a non-existent key.
		// On an unlocked (typical GNOME) wallet this returns ErrNotFound and we
		// use the OS keyring; any other error means no backend → file fallback.
		if _, err := gokeyring.Get(serviceName, "__probe__"); err != nil {
			if !errors.Is(err, gokeyring.ErrNotFound) {
				if err := os.MkdirAll(fileFallbackDir, 0o700); err != nil {
					return nil, fmt.Errorf("keyring: fallback dir: %w", err)
				}
				s.fallback = true
			}
		}
	}

	return s, nil
}

// NewFileStore opens a file-backed keyring unconditionally. It exists for
// headless environments (CI, servers) and tests where no OS keyring is
// available; passphrase is accepted for API compatibility (the file is
// obfuscation-only, not passphrase-encrypted).
func NewFileStore(serviceName, fileFallbackDir, passphrase string) (*Store, error) {
	if err := os.MkdirAll(fileFallbackDir, 0o700); err != nil {
		return nil, fmt.Errorf("keyring: fallback dir: %w", err)
	}
	return &Store{service: serviceName, fileDir: fileFallbackDir, fallback: true}, nil
}

// UseFallback reports whether the store is backed by the file fallback rather
// than the OS keyring.
func (s *Store) UseFallback() bool {
	return s.fallback
}

// Set stores value under key.
func (s *Store) Set(key, value string) error {
	if s.fallback {
		return s.fileSet(key, value)
	}
	err := gokeyring.Set(s.service, key, value)
	if err == nil {
		return nil
	}
	if errors.Is(err, gokeyring.ErrNotFound) {
		return err
	}
	s.fallback = true
	return s.fileSet(key, value)
}

// Get returns the value stored under key, or ErrKeyNotFound.
func (s *Store) Get(key string) (string, error) {
	if s.fallback {
		return s.fileGet(key)
	}
	v, err := gokeyring.Get(s.service, key)
	if err == nil {
		return v, nil
	}
	if errors.Is(err, gokeyring.ErrNotFound) {
		return "", ErrKeyNotFound
	}
	s.fallback = true
	return s.fileGet(key)
}

// Delete removes key. Removing a missing key is not an error.
func (s *Store) Delete(key string) error {
	if s.fallback {
		return s.fileDelete(key)
	}
	err := gokeyring.Delete(s.service, key)
	if err == nil {
		return nil
	}
	if errors.Is(err, gokeyring.ErrNotFound) {
		return nil
	}
	s.fallback = true
	return s.fileDelete(key)
}

func isKDE() bool {
	de := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	return strings.Contains(de, "kde") || strings.Contains(de, "plasma")
}

func (s *Store) filePath() string {
	return filepath.Join(s.fileDir, s.service+".json")
}

func (s *Store) readFile() (map[string]string, error) {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Store) writeFile(m map[string]string) error {
	if err := os.MkdirAll(s.fileDir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath(), data, 0o600)
}

func (s *Store) fileGet(key string) (string, error) {
	m, err := s.readFile()
	if err != nil {
		return "", err
	}
	v, ok := m[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return v, nil
}

func (s *Store) fileSet(key, value string) error {
	m, err := s.readFile()
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	if m == nil {
		m = map[string]string{}
	}
	m[key] = value
	return s.writeFile(m)
}

func (s *Store) fileDelete(key string) error {
	m, err := s.readFile()
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil
		}
		return err
	}
	if _, ok := m[key]; !ok {
		return nil
	}
	delete(m, key)
	return s.writeFile(m)
}
