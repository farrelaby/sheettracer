package db

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Edge struct {
	ID                int64     `db:"id"`
	SourceSpreadsheet int64     `db:"source_spreadsheet"`
	SourceTabID       int64     `db:"source_tab_id"`
	SourceCell        string    `db:"source_cell"`
	TargetGoogleID    string    `db:"target_google_id"`
	TargetRange       *string   `db:"target_range"`
	FirstSeenAt       time.Time `db:"first_seen_at"`
	LastSeenAt        time.Time `db:"last_seen_at"`
}

type EdgeRepo struct {
	db *sqlx.DB
}

func NewEdgeRepo(db *sqlx.DB) *EdgeRepo {
	return &EdgeRepo{db: db}
}

// DeleteBySpreadsheet removes all edges originating from a spreadsheet.
func (r *EdgeRepo) DeleteBySpreadsheet(spreadsheetID int64) error {
	_, err := r.db.Exec(`DELETE FROM edges WHERE source_spreadsheet = ?`, spreadsheetID)
	return err
}

// UpsertAll replaces all edges for a spreadsheet in one transaction.
func (r *EdgeRepo) UpsertAll(spreadsheetID int64, edges []Edge) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	if _, err := tx.Exec(`DELETE FROM edges WHERE source_spreadsheet = ?`, spreadsheetID); err != nil {
		return err
	}

	for _, e := range edges {
		if _, err := tx.Exec(
			`INSERT INTO edges(source_spreadsheet, source_tab_id, source_cell, target_google_id, target_range)
			 VALUES(?, ?, ?, ?, ?)`,
			e.SourceSpreadsheet, e.SourceTabID, e.SourceCell, e.TargetGoogleID, e.TargetRange,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// FanIn returns inbound edge count per target spreadsheet.
func (r *EdgeRepo) FanIn() (map[string]int, error) {
	rows, err := r.db.Query(
		`SELECT target_google_id, COUNT(*) AS fan_in
		 FROM edges GROUP BY target_google_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var googleID string
		var count int
		if err := rows.Scan(&googleID, &count); err != nil {
			return nil, err
		}
		result[googleID] = count
	}
	return result, rows.Err()
}

// BySource returns all edges originating from a spreadsheet.
func (r *EdgeRepo) BySource(spreadsheetID int64) ([]Edge, error) {
	var ee []Edge
	err := r.db.Select(&ee,
		`SELECT id, source_spreadsheet, source_tab_id, source_cell, target_google_id, target_range, first_seen_at, last_seen_at
		 FROM edges WHERE source_spreadsheet = ? ORDER BY source_tab_id, source_cell`,
		spreadsheetID,
	)
	return ee, err
}
