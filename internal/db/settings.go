package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Settings is a key/value repository for non-secret preferences and last app
// state. Secrets (the OAuth refresh token) live in the OS keyring
// (internal/keyring), never here.
//
// Each key corresponds to a settings tab (e.g. "general", "scan"). The value
// is stored as JSONB and can be any JSON-serializable structure.
type Settings struct {
	db *sqlx.DB
}

func NewSettings(db *sqlx.DB) *Settings {
	return &Settings{db: db}
}

// Get returns the raw JSONB value for key and whether it exists.
func (s *Settings) Get(key string) ([]byte, bool, error) {
	var v []byte
	err := s.db.QueryRow(`SELECT v FROM settings WHERE k = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// Set upserts a JSONB value for key. The value is marshaled to JSON before
// storage; pass raw []byte to skip double-marshaling.
func (s *Settings) Set(key string, value any) error {
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("settings: marshal %q: %w", key, err)
		}
		raw = b
	}
	_, err := s.db.Exec(
		`INSERT INTO settings(k, v) VALUES(?, ?)
		 ON CONFLICT(k) DO UPDATE SET v = excluded.v`,
		key, raw,
	)
	return err
}

// Delete removes a key.
func (s *Settings) Delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM settings WHERE k = ?`, key)
	return err
}

// GetJSON unmarshals the stored JSONB value for key into out.
func (s *Settings) GetJSON(key string, out any) (bool, error) {
	raw, ok, err := s.Get(key)
	if err != nil || !ok {
		return ok, err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return true, fmt.Errorf("settings: %q is not valid JSON: %w", key, err)
	}
	return true, nil
}

// SetJSON marshals v and stores it as JSONB.
func (s *Settings) SetJSON(key string, v any) error {
	return s.Set(key, v)
}
