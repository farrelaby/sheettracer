package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrate brings the database at path up to latestSchemaVersion.
//
// Migrations are forward-only and applied in ascending numeric order, each
// inside its own transaction with PRAGMA user_version bumped atomically, so a
// crash can never leave a half-applied migration. Before the first pending
// migration runs, a copy of the database is written alongside the original
// (path + backupSuffix) so an upgrade can be rolled back by restoring it.
//
// If the on-disk version is newer than this binary understands, migrate
// refuses to run rather than risk an old build writing to an unknown schema.
func migrate(handle *sql.DB, path string) error {
	current, err := currentVersion(handle)
	if err != nil {
		return err
	}
	if current > latestSchemaVersion {
		return fmt.Errorf(
			"db: database schema is version %d, but this build of SheetTracer supports up to %d; "+
				"it was created by a newer version of the app — refusing to open it",
			current, latestSchemaVersion,
		)
	}
	if current == latestSchemaVersion {
		return nil
	}

	if err := backup(handle, path); err != nil {
		return err
	}

	for _, n := range pendingMigrations(current) {
		if err := apply(handle, n); err != nil {
			return fmt.Errorf("db: apply migration %04d: %w", n, err)
		}
	}
	return nil
}

func currentVersion(handle *sql.DB) (int, error) {
	var v int
	if err := handle.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return 0, fmt.Errorf("db: read user_version: %w", err)
	}
	return v, nil
}

// pendingMigrations returns the migration numbers > current that this binary
// knows about, in ascending order.
func pendingMigrations(current int) []int {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil
	}
	var nums []int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		n, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0])
		if err != nil || n <= current || n > latestSchemaVersion {
			continue
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)
	return nums
}

// migrationFile resolves migrations/NNN_description.sql for migration number n.
func migrationFile(n int) (string, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return "", err
	}
	prefix := fmt.Sprintf("%04d_", n)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			return path.Join("migrations", e.Name()), nil
		}
	}
	return "", fmt.Errorf("migration %03d not found", n)
}

// apply runs a single migration script and bumps user_version in the same
// transaction.
func apply(handle *sql.DB, n int) error {
	file, err := migrationFile(n)
	if err != nil {
		return err
	}
	content, err := migrationsFS.ReadFile(file)
	if err != nil {
		return err
	}

	tx, err := handle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	if _, err := tx.Exec(string(content)); err != nil {
		return err
	}
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", n)); err != nil {
		return err
	}
	return tx.Commit()
}

// backup writes a single-file copy of the database via VACUUM INTO before the
// first pending migration. VACUUM INTO errors if the destination already
// exists, so the previous backup is replaced. In-memory databases (tests) are
// skipped.
func backup(handle *sql.DB, path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file::memory:") {
		return nil
	}
	dest := path + backupSuffix
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("db: remove stale backup %s: %w", dest, err)
	}
	escaped := strings.ReplaceAll(dest, "'", "''")
	if _, err := handle.Exec("VACUUM INTO '" + escaped + "'"); err != nil {
		return fmt.Errorf("db: backup to %s: %w", dest, err)
	}
	return nil
}