package db

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Spreadsheet struct {
	ID           int64      `db:"id"`
	GoogleID     string     `db:"google_id"`
	Title        string     `db:"title"`
	URL          string     `db:"url"`
	Notes        string     `db:"notes"`
	Version      string     `db:"version"`
	ModifiedTime string     `db:"modified_time"`
	Visibility   string     `db:"visibility"`
	IsTracked    bool       `db:"is_tracked"`
	LastScanAt   *time.Time `db:"last_scan_at"`
	AddedAt      time.Time  `db:"added_at"`
}

type SpreadsheetRepo struct {
	db *sqlx.DB
}

func NewSpreadsheetRepo(db *sqlx.DB) *SpreadsheetRepo {
	return &SpreadsheetRepo{db: db}
}

func (r *SpreadsheetRepo) UpsertByGoogleID(sp *Spreadsheet) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO spreadsheets(google_id, title, url, notes, version, modified_time, visibility)
		 VALUES(?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(google_id) DO UPDATE SET
			title = excluded.title,
			url = excluded.url,
			notes = COALESCE(excluded.notes, spreadsheets.notes),
			version = excluded.version,
			modified_time = excluded.modified_time,
			visibility = excluded.visibility`,
		sp.GoogleID, sp.Title, sp.URL, sp.Notes, sp.Version, sp.ModifiedTime, sp.Visibility,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id == 0 {
		// On conflict (update), LastInsertId returns 0 — fetch the existing id.
		err = r.db.QueryRow(`SELECT id FROM spreadsheets WHERE google_id = ?`, sp.GoogleID).Scan(&id)
	}
	return id, err
}

func (r *SpreadsheetRepo) GetByGoogleID(googleID string) (*Spreadsheet, error) {
	var s Spreadsheet
	err := r.db.QueryRowx(
		`SELECT id, google_id, title, url, notes, version, modified_time, visibility, is_tracked, last_scan_at, added_at
		 FROM spreadsheets WHERE google_id = ?`, googleID,
	).StructScan(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SpreadsheetRepo) ListTracked() ([]Spreadsheet, error) {
	var ss []Spreadsheet
	err := r.db.Select(&ss,
		`SELECT id, google_id, title, url, notes, version, modified_time, visibility, is_tracked, last_scan_at, added_at
		 FROM spreadsheets WHERE is_tracked = 1 ORDER BY added_at DESC`,
	)
	return ss, err
}

func (r *SpreadsheetRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM spreadsheets WHERE id = ?`, id)
	return err
}

func (r *SpreadsheetRepo) UpdateVisibility(id int64, visibility string) error {
	_, err := r.db.Exec(`UPDATE spreadsheets SET visibility = ? WHERE id = ?`, visibility, id)
	return err
}

func (r *SpreadsheetRepo) GetSpreadsheetByID(id int) (Spreadsheet, error) {
	var ss Spreadsheet
	err := r.db.Get(&ss,
		`SELECT id, google_id, title, url, notes, version, modified_time, visibility, is_tracked, last_scan_at, added_at
		 FROM spreadsheets WHERE id=? AND is_tracked = 1 ORDER BY added_at DESC`, id,
	)
	return ss, err
}
