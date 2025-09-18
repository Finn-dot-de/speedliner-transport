package esiauth

import (
	"os"
	"sync"

	"golang.org/x/oauth2"
)

var (
	mu       sync.RWMutex
	memStore = make(map[string]*oauth2.Token) // Fallback, wenn kein DB-Store gesetzt
	store    TokenStore                       // aus store_pg.go
)

func InitStore(s TokenStore) { store = s }

func GetOAuthConfig() *oauth2.Config {
	redirect := os.Getenv("OAUTH_REDIRECT_URL")
	if redirect == "" {
		redirect = "http://localhost:8080/app/callback"
	}
	return &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		Scopes: []string{
			"esi-mail.send_mail.v1",
			"publicData",
		},
		RedirectURL: redirect,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
	}
}

// Persistiert wenn store != nil, sonst In-Memory.
func SaveToken(charID string, token *oauth2.Token) error {
	if store != nil {
		return store.Put(charID, token)
	}
	mu.Lock()
	defer mu.Unlock()
	memStore[charID] = token
	return nil
}

func LoadToken(charID string) (*oauth2.Token, bool) {
	if store != nil {
		return store.Get(charID)
	}
	mu.RLock()
	defer mu.RUnlock()
	tok, ok := memStore[charID]
	return tok, ok
}

func DeleteToken(charID string) error {
	if store != nil {
		return store.Delete(charID)
	}
	mu.Lock()
	delete(memStore, charID)
	mu.Unlock()
	return nil
}
