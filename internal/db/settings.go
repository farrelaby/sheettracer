package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Settings is a key/value repository for non-secret preferences and last app
// state. Secrets (the OAuth refresh token) live in the OS keyring
// (internal/keyring), never here.
type Settings struct {
	db *sql.DB
}

func NewSettings(db *sql.DB) *Settings {
	return &Settings{db: db}
}

// Get returns the stored value for key and whether it exists.
func (s *Settings) Get(key string) (string, bool, error) {
	var v string
	err := s.db.QueryRow(`SELECT v FROM settings WHERE k = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

// Set upserts a value for key.
func (s *Settings) Set(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings(k, v) VALUES(?, ?)
		 ON CONFLICT(k) DO UPDATE SET v = excluded.v, updated_at = datetime('now')`,
		key, value,
	)
	return err
}

// Delete removes a key.
func (s *Settings) Delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM settings WHERE k = ?`, key)
	return err
}

// GetBool returns a stored boolean.
func (s *Settings) GetBool(key string) (bool, bool, error) {
	v, ok, err := s.Get(key)
	if err != nil || !ok {
		return false, ok, err
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, true, fmt.Errorf("settings: %q is not a bool: %w", key, err)
	}
	return b, true, nil
}

// SetBool stores a boolean.
func (s *Settings) SetBool(key string, b bool) error {
	return s.Set(key, strconv.FormatBool(b))
}

// GetInt returns a stored integer.
func (s *Settings) GetInt(key string) (int, bool, error) {
	v, ok, err := s.Get(key)
	if err != nil || !ok {
		return 0, ok, err
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, true, fmt.Errorf("settings: %q is not an int: %w", key, err)
	}
	return n, true, nil
}

// SetInt stores an integer.
func (s *Settings) SetInt(key string, n int) error {
	return s.Set(key, strconv.Itoa(n))
}

// GetJSON unmarshals a stored JSON value into out.
func (s *Settings) GetJSON(key string, out any) (bool, error) {
	v, ok, err := s.Get(key)
	if err != nil || !ok {
		return ok, err
	}
	if err := json.Unmarshal([]byte(v), out); err != nil {
		return true, fmt.Errorf("settings: %q is not valid JSON: %w", key, err)
	}
	return true, nil
}

// SetJSON marshals v and stores it as a JSON value.
func (s *Settings) SetJSON(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Set(key, string(b))
}