package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sheettracer/internal/oauth"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// AuthState is mirrored to the frontend via the "auth:state" event.
type AuthState struct {
	Connected bool   `json:"connected"`
	Status    string `json:"status"` // disconnected | connecting | connected | expired
	Message   string `json:"message,omitempty"`
}

const authStateEvent = "auth:state"

// OAuthService exposes Google account connection to the frontend. It is a
// Wails-bound service: the only place that touches app events / the browser.
type OAuthService struct {
	client  *oauth.Client
	store   oauth.TokenStore
	openURL func(string) error
	emit    func(event string, data any)

	connecting sync.Mutex
}

func NewOAuthService(client *oauth.Client, store oauth.TokenStore,
	openURL func(string) error, emit func(string, any)) *OAuthService {
	return &OAuthService{client: client, store: store, openURL: openURL, emit: emit}
}

// Status reports the current connection state. It validates and silently
// refreshes a stored token so a returning user reconnects without the browser.
func (s *OAuthService) Status() (*AuthState, error) {
	tok, err := s.store.Get()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return &AuthState{Status: "disconnected"}, nil
	}
	if tok.RefreshToken == "" {
		return &AuthState{Status: "expired"}, nil
	}
	refreshed, err := s.client.Source(context.Background(), tok, s.store).Token()
	if err != nil {
		return &AuthState{Status: "expired", Message: err.Error()}, nil
	}
	if !refreshed.Expiry.IsZero() && time.Until(refreshed.Expiry) <= 0 {
		return &AuthState{Status: "expired"}, nil
	}
	return &AuthState{Connected: true, Status: "connected"}, nil
}

// Connect starts the browser consent flow. It returns immediately; the
// outcome arrives through the "auth:state" event. Concurrent calls while a
// flow is in progress are ignored and just echo the connecting state.
func (s *OAuthService) Connect() *AuthState {
	if !s.connecting.TryLock() {
		return &AuthState{Status: "connecting"}
	}

	s.emit(authStateEvent, AuthState{Status: "connecting"})
	go func() {
		defer s.connecting.Unlock()
		tok, err := s.client.Connect(context.Background(), s.openURL)
		if err != nil {
			s.emit(authStateEvent, AuthState{Status: "disconnected", Message: err.Error()})
			return
		}
		if err := s.store.Save(tok); err != nil {
			s.emit(authStateEvent, AuthState{Status: "disconnected", Message: err.Error()})
			return
		}
		s.emit(authStateEvent, AuthState{Connected: true, Status: "connected"})
	}()
	return &AuthState{Status: "connecting"}
}

// Disconnect removes the stored tokens.
func (s *OAuthService) Disconnect() error {
	if err := s.store.Delete(); err != nil {
		return err
	}
	s.emit(authStateEvent, AuthState{Status: "disconnected"})
	return nil
}

type AccountInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	PhotoUrl string `json:"photo_url"`
}

// Returns user account info (name, email)
func (s *OAuthService) Account() (AccountInfo, error) {
	tok, err := s.store.Get()
	if err != nil {
		return AccountInfo{}, err
	}
	if tok == nil {
		return AccountInfo{}, fmt.Errorf("No token found")
	}

	src := s.client.Source(context.Background(), tok, s.store)

	driveService, err := drive.NewService(context.Background(), option.WithTokenSource(src))
	if err != nil {
		return AccountInfo{}, fmt.Errorf("[drive Service init]: %w", err)
	}
	about, err := driveService.About.Get().Fields("user").Do()
	if err != nil {

		return AccountInfo{}, fmt.Errorf("[about.get]: %w", err)
	}

	return AccountInfo{
		Name:  about.User.DisplayName,
		Email: about.User.EmailAddress,
		PhotoUrl: about.User.PhotoLink,
	}, nil

}
