package db

import "github.com/jmoiron/sqlx"

type Tab struct {
	ID            int64  `db:"id"`
	SpreadsheetID int64  `db:"spreadsheet_id"`
	TabID         int64  `db:"tab_id"`
	Title         string `db:"title"`
	Idx           int    `db:"idx"`
}

type TabRepo struct {
	db *sqlx.DB
}

func NewTabRepo(db *sqlx.DB) *TabRepo {
	return &TabRepo{db: db}
}

// UpsertAll replaces all tabs for a spreadsheet in one transaction.
func (r *TabRepo) UpsertAll(spreadsheetID int64, tabs []Tab) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	if _, err := tx.Exec(`DELETE FROM tabs WHERE spreadsheet_id = ?`, spreadsheetID); err != nil {
		return err
	}

	for _, t := range tabs {
		if _, err := tx.Exec(
			`INSERT INTO tabs(spreadsheet_id, tab_id, title, idx) VALUES(?, ?, ?, ?)`,
			spreadsheetID, t.TabID, t.Title, t.Idx,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *TabRepo) ListBySpreadsheet(spreadsheetID int64) ([]Tab, error) {
	var tt []Tab
	err := r.db.Select(&tt,
		`SELECT id, spreadsheet_id, tab_id, title, idx
		 FROM tabs WHERE spreadsheet_id = ? ORDER BY idx`,
		spreadsheetID,
	)
	return tt, err
}
