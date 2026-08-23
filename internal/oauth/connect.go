package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	callbackPath    = "/oauth/callback"
	loopbackTimeout = 5 * time.Minute
)

var errConsentDenied = errors.New("oauth: consent denied")

// Connect runs the desktop authorization-code flow:
//
//  1. openURL(authURL) opens Google's consent page in the system browser.
//  2. The user approves; Google redirects the browser to our loopback server
//     with an authorization code.
//  3. We validate the state parameter, then exchange the code for tokens.
//
// openURL is invoked with the consent URL and must open it in a browser.
func (c *Client) Connect(ctx context.Context, openURL func(string) error) (*oauth2.Token, error) {
	state, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	verifier := oauth2.GenerateVerifier()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("oauth: loopback listener: %w", err)
	}
	defer l.Close()

	cfg := c.config
	cfg.RedirectURL = "http://" + l.Addr().String() + callbackPath
	authURL := cfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,             // request a refresh token
		oauth2.S256ChallengeOption(verifier), // PKCE
	)

	codeCh := make(chan codeResult, 1)
	srv := &http.Server{Handler: c.callbackHandler(state, codeCh)}
	go srv.Serve(l) //nolint:errcheck
	defer srv.Close()

	if err := openURL(authURL); err != nil {
		return nil, fmt.Errorf("oauth: open browser: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, loopbackTimeout)
	defer cancel()

	var res codeResult
	select {
	case res = <-codeCh:
	case <-ctx.Done():
		return nil, fmt.Errorf("oauth: timed out waiting for approval: %w", ctx.Err())
	}
	if res.err != nil {
		return nil, res.err
	}

	tok, err := cfg.Exchange(ctx, res.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("oauth: token exchange: %w", err)
	}
	return tok, nil
}

type codeResult struct {
	code string
	err  error
}

func (c *Client) callbackHandler(state string, out chan<- codeResult) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("state"); got != state {
			out <- codeResult{err: errors.New("oauth: state mismatch (possible CSRF)")}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		if desc := q.Get("error"); desc != "" {
			out <- codeResult{err: fmt.Errorf("%w: %s", errConsentDenied, desc)}
			renderPage(w, "Authorization failed", "You can close this tab and return to SheetTracer.")
			return
		}
		code := q.Get("code")
		if code == "" {
			out <- codeResult{err: errors.New("oauth: missing authorization code")}
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			return
		}
		out <- codeResult{code: code}
		renderPage(w, "Authorization complete", "You can close this tab and return to SheetTracer.")
	})
	return mux
}

func renderPage(w http.ResponseWriter, title, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>%s</title></head>"+
		"<body style=\"font-family:system-ui,-apple-system,sans-serif;text-align:center;padding-top:12vh\">"+
		"<h1>%s</h1><p>%s</p></body></html>", title, title, body)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
