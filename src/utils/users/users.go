package users

import (
	"context"
	"fmt"
	"speedliner-server/db"
)

func GetUserRole(charID int64) (string, error) {
	var role string
	err := db.Pool.QueryRow(context.Background(), `
		SELECT role FROM users WHERE char_id = $1
	`, charID).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("GetUserRole error: %w", err)
	}
	return role, nil
}
