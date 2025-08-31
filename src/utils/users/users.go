package users

import (
	"context"
	"fmt"
	"speedliner-server/db"
)

func HasRole(charID string, allowedRoles ...string) (bool, error) {
	query := `
		SELECT role 
		FROM users 
		WHERE char_id = $1;`

	row := db.Pool.QueryRow(context.Background(), query, charID)
	var userRole string
	if err := row.Scan(&userRole); err != nil {
		return false, fmt.Errorf("GetUserRoles error: %w", err)
	}

	for _, role := range allowedRoles {
		if userRole == role {
			return true, nil
		}
	}
	return false, nil
}
