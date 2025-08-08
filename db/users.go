package db

import (
	"context"
	"fmt"
)

func UpsertUser(charID string, name string) error {
	query := `
	INSERT INTO users (char_id, name, role)
	VALUES ($1, $2, 'user')
	ON CONFLICT (char_id) DO UPDATE
	SET name = EXCLUDED.name;
	`

	_, err := Pool.Exec(context.Background(), query, charID, name)
	if err != nil {
		return fmt.Errorf("UpsertUser error: %w", err)
	}
	return nil
}

func GetUserRoles(charID string) (string, error) {
	query := `
		SELECT role 
		FROM users 
		WHERE char_id = $1;`

	row := Pool.QueryRow(context.Background(), query, charID)
	var role string
	if err := row.Scan(&role); err != nil {
		return "", fmt.Errorf("GetUserRoles error: %w", err)
	}
	return role, nil
}
