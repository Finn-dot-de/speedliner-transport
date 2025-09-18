// esiauth/store.go
package esiauth

import (
	"database/sql"
	"encoding/json"
	"time"

	"golang.org/x/oauth2"
)

type TokenStore interface {
	Get(charID string) (*oauth2.Token, bool)
	Put(charID string, tok *oauth2.Token) error
	Delete(charID string) error
}

type DBTokenStore struct{ DB *sql.DB }

func (s *DBTokenStore) Get(charID string) (*oauth2.Token, bool) {
	var js string
	err := s.DB.QueryRow(`SELECT token_json FROM oauth_tokens WHERE char_id=$1`, charID).Scan(&js)
	if err != nil {
		return nil, false
	}
	var tok oauth2.Token
	if json.Unmarshal([]byte(js), &tok) != nil {
		return nil, false
	}
	return &tok, true
}

func (s *DBTokenStore) Put(charID string, tok *oauth2.Token) error {
	b, _ := json.Marshal(tok) // Optional: hier verschlüsseln (AES-GCM) mit KEY aus ENV
	_, err := s.DB.Exec(`
		INSERT INTO oauth_tokens (char_id, token_json, updated_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (char_id) DO UPDATE SET token_json=EXCLUDED.token_json, updated_at=EXCLUDED.updated_at
	`, charID, string(b), time.Now())
	return err
}

func (s *DBTokenStore) Delete(charID string) error {
	_, err := s.DB.Exec(`DELETE FROM oauth_tokens WHERE char_id=$1`, charID)
	return err
}
