package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T) (path string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "test.db")
	handle, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { handle.Close() })
	return path
}

func mustReopen(t *testing.T, path string) *sql.DB {
	t.Helper()
	handle, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { handle.Close() })
	return handle
}

func TestOpenCreatesFullSchema(t *testing.T) {
	path := openTestDB(t)
	handle := mustReopen(t, path)

	var v int
	if err := handle.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != latestSchemaVersion {
		t.Fatalf("user_version = %d, want %d", v, latestSchemaVersion)
	}

	for _, table := range []string{"settings", "spreadsheets", "sheets", "imports", "scan_cache", "scan_runs"} {
		var name string
		err := handle.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("db perms = %o, want 600", perm)
	}
}

func TestReopenIsNoOp(t *testing.T) {
	path := openTestDB(t)
	before, err := os.Stat(path + backupSuffix)
	if err != nil {
		t.Fatalf("initial upgrade should have created a backup: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	mustReopen(t, path)

	after, err := os.Stat(path + backupSuffix)
	if err != nil {
		t.Fatalf("backup missing after reopen: %v", err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("no-op reopen should not rewrite the backup")
	}
}

func TestRefusesNewerSchema(t *testing.T) {
	path := openTestDB(t)
	handle := mustReopen(t, path)
	if _, err := handle.Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(path); err == nil {
		t.Fatal("expected error opening a database with a newer schema")
	}
}

func TestBackupCreatedBeforeUpgrade(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	handle, err := Open(path)
	if err != nil {
		t.Fatalf("Open on fresh v0 file: %v", err)
	}
	t.Cleanup(func() { handle.Close() })

	if _, err := os.Stat(path + backupSuffix); err != nil {
		t.Fatalf("pre-upgrade backup missing: %v", err)
	}
}

func TestMigrationFileResolution(t *testing.T) {
	file, err := migrationFile(1)
	if err != nil {
		t.Fatal(err)
	}
	if file != "migrations/0001_init.sql" {
		t.Fatalf("resolved %q", file)
	}
	if _, err := migrationFile(99); err == nil {
		t.Fatal("expected error for unknown migration number")
	}
}