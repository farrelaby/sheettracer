package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "turso.tech/database/tursogo"
)

// latestSchemaVersion is the newest migration this binary understands. Bump it
// (and add a numbered .sql file under migrations/) whenever the schema changes.
const latestSchemaVersion = 1

// backupSuffix is appended to the database path for the pre-upgrade copy.
const backupSuffix = ".pre-upgrade.db"

// Open opens (creating if needed) the Turso/libSQL database at path, applies
// any pending migrations, and returns a ready *sql.DB.
//
// The data directory is created 0700 and the database file is chmod'ed 0600 so
// the credentials SheetTracer persists stay private to the owning user.
func Open(path string) (*sqlx.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("db: create data dir: %w", err)
		}
	}

	handle, err := sqlx.Open("turso", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}

	if err := setPragmas(handle); err != nil {
		handle.Close()
		return nil, err
	}

	if err := handle.Ping(); err != nil {
		handle.Close()
		return nil, fmt.Errorf("db: ping %s: %w", path, err)
	}

	if err := migrate(handle.DB, path); err != nil {
		handle.Close()
		return nil, err
	}

	if err := os.Chmod(path, 0o600); err != nil {
		handle.Close()
		return nil, fmt.Errorf("db: chmod %s: %w", path, err)
	}

	return handle, nil
}

// setPragmas configures the connection for WAL mode, foreign keys, and
// performance-tuned synchronous writes. Each PRAGMA is executed on the
// connection individually (foreign_keys is per-connection by design).
func setPragmas(handle *sqlx.DB) error {
	pragmas := []string{
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
	}
	for _, p := range pragmas {
		if _, err := handle.Exec(p); err != nil {
			return fmt.Errorf("db: %s: %w", p, err)
		}
	}
	return nil
}