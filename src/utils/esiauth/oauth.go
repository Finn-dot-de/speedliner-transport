package esiauth

import (
	"os"
	"sync"

	"golang.org/x/oauth2"
)

var (
	tokenStore = make(map[string]*oauth2.Token)
	mu         sync.RWMutex
)

func GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		Scopes:       []string{},
		RedirectURL:  "http://localhost:8080/app/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
	}
}

func SaveToken(charID string, token *oauth2.Token) {
	mu.Lock()
	defer mu.Unlock()
	tokenStore[charID] = token
}

func LoadToken(charID string) (*oauth2.Token, bool) {
	mu.RLock()
	defer mu.RUnlock()
	token, ok := tokenStore[charID]
	return token, ok
}

func DeleteToken(charID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(tokenStore, charID)
}
