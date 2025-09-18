package esiauth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

type PGXTokenStore struct{ Pool *pgxpool.Pool }

func NewPGXTokenStore(pool *pgxpool.Pool) *PGXTokenStore {
	return &PGXTokenStore{Pool: pool}
}

func (s *PGXTokenStore) Get(charID string) (*oauth2.Token, bool) {
	var js string
	err := s.Pool.QueryRow(context.Background(),
		`SELECT token_json FROM oauth_tokens WHERE char_id=$1`, charID).
		Scan(&js)
	if err != nil {
		return nil, false
	}
	var tok oauth2.Token
	if json.Unmarshal([]byte(js), &tok) != nil {
		return nil, false
	}
	return &tok, true
}

func (s *PGXTokenStore) Put(charID string, tok *oauth2.Token) error {
	b, _ := json.Marshal(tok) // optional: verschlüsseln (AES-GCM)
	_, err := s.Pool.Exec(context.Background(),
		`INSERT INTO oauth_tokens (char_id, token_json, updated_at)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (char_id) DO UPDATE
		   SET token_json=EXCLUDED.token_json, updated_at=EXCLUDED.updated_at`,
		charID, string(b), time.Now())
	return err
}

func (s *PGXTokenStore) Delete(charID string) error {
	_, err := s.Pool.Exec(context.Background(),
		`DELETE FROM oauth_tokens WHERE char_id=$1`, charID)
	return err
}
