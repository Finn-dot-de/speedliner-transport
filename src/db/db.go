package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("DB connect error: %w", err)
	}
	if err := Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("DB ping error: %w", err)
	}
	fmt.Println("✅ DB connected")

	if err := ensureSchema(); err != nil {
		return fmt.Errorf("DB schema setup error: %w", err)
	}
	return nil
}

func ensureSchema() error {
	ctx := context.Background()
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,

		`CREATE TABLE IF NOT EXISTS alliances (
			alliance_id BIGINT PRIMARY KEY,
			name        TEXT NOT NULL,
			ticker      TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS corps (
			corp_id     BIGINT PRIMARY KEY,
			name        TEXT NOT NULL,
			ticker      TEXT,
			alliance_id BIGINT NULL REFERENCES alliances(alliance_id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS users (
			char_id BIGINT PRIMARY KEY,
			name    TEXT NOT NULL,
			role    TEXT NOT NULL DEFAULT 'user',
			corp_id BIGINT NULL REFERENCES corps(corp_id) ON DELETE SET NULL
		);`,

		// -> Check-Constraint gleich im CREATE (greift bei frischer DB)
		`CREATE TABLE IF NOT EXISTS routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			from_system   TEXT NOT NULL,
			to_system     TEXT NOT NULL,
			price_per_m3  NUMERIC(10,2),
			no_collateral BOOLEAN NOT NULL DEFAULT false,
			visibility    TEXT NOT NULL DEFAULT 'all',
			CONSTRAINT routes_visibility_chk CHECK (visibility IN ('all','whitelist'))
		);`,

		`CREATE INDEX IF NOT EXISTS idx_users_corp_id  ON users(corp_id);`,
		`CREATE INDEX IF NOT EXISTS idx_corps_alliance ON corps(alliance_id);`,

		`CREATE OR REPLACE VIEW v_users_enriched AS
		 SELECT u.char_id, u.name, u.role,
		        u.corp_id,       c.name  AS corp_name,     c.ticker  AS corp_ticker,
		        c.alliance_id,   a.name  AS alliance_name, a.ticker  AS alliance_ticker
		 FROM users u
		 LEFT JOIN corps     c ON c.corp_id = u.corp_id
		 LEFT JOIN alliances a ON a.alliance_id = c.alliance_id;`,

		`CREATE TABLE IF NOT EXISTS oauth_tokens (
			  char_id TEXT PRIMARY KEY,
			  token_json TEXT NOT NULL,
			  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			`,
		 
		`DO $$
		BEGIN
		  IF NOT EXISTS (
		    SELECT 1
		    FROM   pg_constraint
		    WHERE  conname = 'routes_visibility_chk'
		    AND    conrelid = 'routes'::regclass
		  ) THEN
		    ALTER TABLE routes
		      ADD CONSTRAINT routes_visibility_chk
		      CHECK (visibility IN ('all','whitelist'));
		  END IF;
		END$$;`,

		`CREATE TABLE IF NOT EXISTS route_visibility (
		  route_id UUID  NOT NULL REFERENCES routes(id)  ON DELETE CASCADE,
		  corp_id  BIGINT NOT NULL REFERENCES corps(corp_id) ON DELETE CASCADE,
		  PRIMARY KEY (route_id, corp_id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_route_visibility_corp ON route_visibility(corp_id);`,
	}

	for _, s := range stmts {
		if _, err = tx.Exec(ctx, s); err != nil {
			return err
		}
	}
	fmt.Println("✅ DB schema checked/created")
	return nil
}
