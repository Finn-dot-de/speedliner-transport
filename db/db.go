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

	// Tabellen prüfen / anlegen
	if err := ensureSchema(); err != nil {
		return fmt.Errorf("DB schema setup error: %w", err)
	}

	return nil
}

func ensureSchema() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,

		`CREATE TABLE IF NOT EXISTS users (
			char_id BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user'
		);`,

		`CREATE TABLE IF NOT EXISTS routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			from_system TEXT NOT NULL,
			to_system TEXT NOT NULL,
			price_per_m3 NUMERIC(10, 2)
		);`,
	}

	for _, q := range queries {
		if _, err := Pool.Exec(context.Background(), q); err != nil {
			return err
		}
	}

	fmt.Println("✅ DB schema checked/created")
	return nil
}
