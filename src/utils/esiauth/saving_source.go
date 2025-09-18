package esiauth

import "golang.org/x/oauth2"

// Wrappt einen TokenSource und speichert JEDE Erneuerung sofort (Refresh-Token-Rotation).
type savingTokenSource struct {
	charID string
	inner  oauth2.TokenSource
}

func NewSavingTokenSource(charID string, inner oauth2.TokenSource) oauth2.TokenSource {
	return &savingTokenSource{charID: charID, inner: inner}
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.inner.Token()
	if err == nil && t != nil {
		_ = SaveToken(s.charID, t) // bewusst ignoriert; optional loggen
	}
	return t, err
}
