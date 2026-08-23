package oauth

import (
	"golang.org/x/oauth2"
)

// persistingSource wraps the library token source so every real refresh
// round-trips through the keyring. Valid tokens short-circuit inside the
// library's ReuseTokenSource, so this only writes on an actual refresh.
type persistingSource struct {
	src   oauth2.TokenSource
	store TokenStore
}

func (p *persistingSource) Token() (*oauth2.Token, error) {
	tok, err := p.src.Token()
	if err != nil {
		return nil, err
	}
	if err := p.store.Save(tok); err != nil {
		return nil, err
	}
	return tok, nil
}
