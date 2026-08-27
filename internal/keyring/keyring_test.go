package keyring

import (
	"errors"
	"testing"
)

// newFileTestStore exercises the file backend used by the fallback path.
func newFileTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewFileStore("SheetTracer-test", t.TempDir(), fallbackPassphrase)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	return s
}

func TestSetGetDelete(t *testing.T) {
	s := newFileTestStore(t)

	if err := s.Set(RefreshToken, "token-123"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(RefreshToken)
	if err != nil || got != "token-123" {
		t.Fatalf("Get: got %q err=%v", got, err)
	}

	if err := s.Delete(RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(RefreshToken); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("want ErrKeyNotFound after delete, got %v", err)
	}
}

func TestGetMissingKey(t *testing.T) {
	s := newFileTestStore(t)
	if _, err := s.Get("does-not-exist"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("want ErrKeyNotFound, got %v", err)
	}
}

func TestOverwrite(t *testing.T) {
	s := newFileTestStore(t)
	if err := s.Set(TokenJSON, `{"a":1}`); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(TokenJSON, `{"a":2}`); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(TokenJSON)
	if err != nil || got != `{"a":2}` {
		t.Fatalf("overwrite: got %q err=%v", got, err)
	}
}

func TestFallbackFlag(t *testing.T) {
	s := newFileTestStore(t)
	if !s.UseFallback() {
		t.Fatal("file store should report UseFallback() == true")
	}
}
