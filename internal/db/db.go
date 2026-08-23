package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// latestSchemaVersion is the newest migration this binary understands. Bump it
// (and add a numbered .sql file under migrations/) whenever the schema changes.
const latestSchemaVersion = 1

// backupSuffix is appended to the database path for the pre-upgrade copy.
const backupSuffix = ".pre-upgrade.db"

// Open opens (creating if needed) the SQLite database at path, applies any
// pending migrations, and returns a ready *sql.DB.
//
// The data directory is created 0700 and the database file is chmod'ed 0600 so
// the credentials SheetTracer persists stay private to the owning user.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("db: create data dir: %w", err)
		}
	}

	handle, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	if err := handle.Ping(); err != nil {
		handle.Close()
		return nil, fmt.Errorf("db: ping %s: %w", path, err)
	}

	if err := migrate(handle, path); err != nil {
		handle.Close()
		return nil, err
	}

	if err := os.Chmod(path, 0o600); err != nil {
		handle.Close()
		return nil, fmt.Errorf("db: chmod %s: %w", path, err)
	}

	return handle, nil
}

// dsn builds a SQLite URI. Pragmas are passed as _pragma parameters so they are
// applied to every pooled connection (PRAGMA statements executed via Exec only
// affect the one connection they run on, and foreign_keys is per-connection).
func dsn(path string) string {
	escaped := strings.NewReplacer(" ", "%20", "#", "%23", "?", "%3F").Replace(path)
	return "file:" + escaped +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=synchronous(NORMAL)"
}