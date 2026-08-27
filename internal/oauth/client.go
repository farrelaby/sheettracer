package oauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Scopes requested at consent time. Both are read-only.
var Scopes = []string{
	"https://www.googleapis.com/auth/spreadsheets.readonly",
	"https://www.googleapis.com/auth/drive.metadata.readonly",
}

// Client drives the authorization-code + PKCE flow against Google.
type Client struct {
	config oauth2.Config
}

// New returns a Client for a Google Cloud *Desktop* OAuth client.
func New(clientID, clientSecret string) *Client {
	return &Client{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       Scopes,
		},
	}
}

// Source returns a token source that reuses tok, silently refreshes it when
// expired, and persists each refreshed token through store. Valid tokens
// short-circuit inside the library's ReuseTokenSource, so this only writes
// on an actual refresh.
func (c *Client) Source(ctx context.Context, tok *oauth2.Token, store TokenStore) oauth2.TokenSource {
	return &persistingSource{src: c.config.TokenSource(ctx, tok), store: store, lastSaved: tok}
}
