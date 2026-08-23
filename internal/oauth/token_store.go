package oauth

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/oauth2"

	"sheettracer/internal/keyring"
)

// TokenStore persists the full token payload outside SQLite.
type TokenStore interface {
	Save(*oauth2.Token) error
	Get() (*oauth2.Token, error)
	Delete() error
}

// KeyringStore stores tokens in the OS keyring via internal/keyring.
type KeyringStore struct {
	kr *keyring.Store
}

// NewKeyringStore wraps an internal/keyring store.
func NewKeyringStore(kr *keyring.Store) *KeyringStore {
	return &KeyringStore{kr: kr}
}

func (s *KeyringStore) Save(tok *oauth2.Token) error {
	raw, err := json.Marshal(tok)
	if err != nil {
		return fmt.Errorf("oauth: marshal token: %w", err)
	}
	if err := s.kr.Set(keyring.TokenJSON, string(raw)); err != nil {
		return fmt.Errorf("oauth: keyring set: %w", err)
	}
	if tok.RefreshToken != "" {
		if err := s.kr.Set(keyring.RefreshToken, tok.RefreshToken); err != nil {
			return fmt.Errorf("oauth: keyring set refresh token: %w", err)
		}
	}
	return nil
}

func (s *KeyringStore) Get() (*oauth2.Token, error) {
	raw, err := s.kr.Get(keyring.TokenJSON)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, nil // not connected
	}
	if err != nil {
		return nil, fmt.Errorf("oauth: keyring get: %w", err)
	}
	tok := new(oauth2.Token)
	if err := json.Unmarshal([]byte(raw), tok); err != nil {
		return nil, fmt.Errorf("oauth: unmarshal token: %w", err)
	}
	return tok, nil
}

func (s *KeyringStore) Delete() error {
	for _, k := range []string{keyring.TokenJSON, keyring.RefreshToken} {
		if err := s.kr.Delete(k); err != nil {
			return fmt.Errorf("oauth: keyring delete %s: %w", k, err)
		}
	}
	return nil
}
