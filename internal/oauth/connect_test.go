package oauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// fakeGoogle returns a token endpoint that issues a fixed token for any
// authorization code, and an auth endpoint that is never actually hit (the
// test drives the callback directly).
func fakeTokenServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"ACCESS","refresh_token":"REFRESH","expires_in":3600,"token_type":"Bearer"}`))
	})
	return httptest.NewServer(mux)
}

// TestConnectLoopback exercises the full desktop flow end-to-end against a fake
// token server. The openURL callback acts as "Google" by replaying the
// redirect to our loopback listener with a code + the original state.
func TestConnectLoopback(t *testing.T) {
	tokenSrv := fakeTokenServer()
	defer tokenSrv.Close()

	c := &Client{
		config: oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Scopes:       Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:   tokenSrv.URL + "/auth",
				TokenURL:  tokenSrv.URL + "/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
	}

	openURL := func(authURL string) error {
		q, err := url.ParseQuery(strings.SplitN(authURL, "?", 2)[1])
		if err != nil {
			return err
		}
		redirectURI := q.Get("redirect_uri")
		state := q.Get("state")
		cb := redirectURI + "?code=TESTCODE&state=" + state
		resp, err := http.Get(cb) //nolint:gosec // loopback test URL
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		return nil
	}

	tok, err := c.Connect(context.Background(), openURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if tok == nil || tok.AccessToken != "ACCESS" || tok.RefreshToken != "REFRESH" {
		t.Fatalf("unexpected token: %+v", tok)
	}
}

// TestConnectStateMismatch ensures a CSRF-mismatched callback is rejected and
// no token is exchanged.
func TestConnectStateMismatch(t *testing.T) {
	tokenSrv := fakeTokenServer()
	defer tokenSrv.Close()

	c := &Client{
		config: oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL:  tokenSrv.URL + "/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
	}

	openURL := func(authURL string) error {
		q, _ := url.ParseQuery(strings.SplitN(authURL, "?", 2)[1])
		cb := q.Get("redirect_uri") + "?code=TESTCODE&state=WRONGSTATE"
		resp, err := http.Get(cb) //nolint:gosec // loopback test URL
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		return nil
	}

	if _, err := c.Connect(context.Background(), openURL); err == nil {
		t.Fatal("expected error on state mismatch")
	}
}
