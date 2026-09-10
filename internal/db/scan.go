package db

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// TriggeredBy records what kicked off a scan run.
// CHECK (manual, launch, add) in migrations/0001_init.sql.
type TriggeredBy string

const (
	TriggerManual TriggeredBy = "manual"
	TriggerLaunch TriggeredBy = "launch"
	TriggerAdd    TriggeredBy = "add"
)

// ScanStatus — CHECK (running, ok, partial, error)
type ScanStatus string

const (
	ScanStatusRunning ScanStatus = "running"
	ScanStatusOK      ScanStatus = "ok"
	ScanStatusPartial ScanStatus = "partial"
	ScanStatusError   ScanStatus = "error"
)

type ScanRun struct {
	ID          int64       `db:"id"`
	TriggeredBy TriggeredBy `db:"triggered_by"`
	StartedAt   time.Time   `db:"started_at"`
	FinishedAt  *time.Time  `db:"finished_at"` // NULL until run completes
	Status      ScanStatus  `db:"status"`
	TabsScanned int         `db:"tabs_scanned"`
	TabsSkipped int         `db:"tabs_skipped"`
	Error       *string     `db:"error"` // NULL when no error
}

type ScanCache struct {
	ID            int64     `db:"id"`
	SpreadsheetID int64     `db:"spreadsheet_id"`
	TabID         *int64    `db:"tab_id"`      // NULLABLE in schema; note UNIQUE(spreadsheet_id, tab_id) treats NULLs as distinct in SQLite
	Payload       []byte    `db:"payload"`     // BLOB NOT NULL (compressed payload per docs/DATABASE.md)
	Fingerprint   *string   `db:"fingerprint"` // NULLABLE TEXT
	FetchedAt     time.Time `db:"fetched_at"`  // NOT NULL DEFAULT CURRENT_TIMESTAMP
}

type ScanRunRepo struct{ db *sqlx.DB }
type ScanCacheRepo struct{ db *sqlx.DB }

func NewScanRunRepo(db *sqlx.DB) *ScanRunRepo {
	return &ScanRunRepo{db: db}
}

func NewScanCacheRepo(db *sqlx.DB) *ScanCacheRepo {
	return &ScanCacheRepo{db: db}
}

func (r *ScanRunRepo) Add(trigger TriggeredBy, start time.Time) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO scan_runs(triggered_by, started_at, status) VALUES(?, ?, ?)`,
		trigger, start, ScanStatusRunning,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Update marks a run finished with its outcome. errMsg is stored in the
// error column (NULL when the run has no error).
func (r *ScanRunRepo) Update(id int64, finish time.Time, status ScanStatus, tabsScanned, tabsSkipped int, errMsg *string) error {
	_, err := r.db.Exec(
		`UPDATE scan_runs SET finished_at = ?, status = ?, tabs_scanned = ?, tabs_skipped = ?, error = ? WHERE id = ?`,
		finish, status, tabsScanned, tabsSkipped, errMsg, id,
	)

	return err
}
