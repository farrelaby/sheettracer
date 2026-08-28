package db

import "github.com/jmoiron/sqlx"

type Sheet struct {
	ID            int64  `db:"id"`
	SpreadsheetID int64  `db:"spreadsheet_id"`
	TabID         int64  `db:"tab_id"`
	Title         string `db:"title"`
	Idx           int    `db:"idx"`
}

type SheetRepo struct {
	db *sqlx.DB
}

func NewSheetRepo(db *sqlx.DB) *SheetRepo {
	return &SheetRepo{db: db}
}

// UpsertAll replaces all tabs for a spreadsheet in one transaction.
func (r *SheetRepo) UpsertAll(spreadsheetID int64, tabs []Sheet) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	if _, err := tx.Exec(`DELETE FROM sheets WHERE spreadsheet_id = ?`, spreadsheetID); err != nil {
		return err
	}

	for _, t := range tabs {
		if _, err := tx.Exec(
			`INSERT INTO sheets(spreadsheet_id, tab_id, title, idx) VALUES(?, ?, ?, ?)`,
			spreadsheetID, t.TabID, t.Title, t.Idx,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *SheetRepo) ListBySpreadsheet(spreadsheetID int64) ([]Sheet, error) {
	var ss []Sheet
	err := r.db.Select(&ss,
		`SELECT id, spreadsheet_id, tab_id, title, idx
		 FROM sheets WHERE spreadsheet_id = ? ORDER BY idx`,
		spreadsheetID,
	)
	return ss, err
}
