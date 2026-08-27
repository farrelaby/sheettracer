package oauth

import (
	"testing"

	"golang.org/x/oauth2"

	"sheettracer/internal/keyring"
)

func newTestStore(t *testing.T) *keyring.Store {
	t.Helper()
	dir := t.TempDir()
	kr, err := keyring.NewFileStore("sheettracer-test", dir, "test-passphrase")
	if err != nil {
		t.Fatalf("keyring store: %v", err)
	}
	return kr
}

func TestKeyringStoreRoundTrip(t *testing.T) {
	store := NewKeyringStore(newTestStore(t))

	tok := &oauth2.Token{
		AccessToken:  "at",
		RefreshToken: "rt",
		TokenType:    "Bearer",
	}
	if err := store.Save(tok); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.AccessToken != "at" || got.RefreshToken != "rt" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, err = store.Get()
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestKeyringStoreMissingIsNil(t *testing.T) {
	got, err := NewKeyringStore(newTestStore(t)).Get()
	if err != nil {
		t.Fatalf("get on empty: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}
