package db

import (
	"context"
	"fmt"
	"speedliner-server/src/utils/structs"
)

func UpsertUser(charID string, name string) error {
	query := `
    INSERT INTO users (char_id, name)
    VALUES ($1, $2)
    ON CONFLICT (char_id) DO UPDATE
    SET name = EXCLUDED.name;`
	_, err := Pool.Exec(context.Background(), query, charID, name)
	if err != nil {
		return fmt.Errorf("UpsertUser error: %w", err)
	}
	return nil
}

func GetUserRoles(charID string) (string, error) {
	query := `SELECT role FROM users WHERE char_id = $1;`
	row := Pool.QueryRow(context.Background(), query, charID)
	var role string
	if err := row.Scan(&role); err != nil {
		return "", fmt.Errorf("GetUserRoles error: %w", err)
	}
	return role, nil
}

func UpdateUserRole(charID string, role string) error {
	query := `UPDATE users SET role = $1 WHERE char_id = $2;`
	_, err := Pool.Exec(context.Background(), query, role, charID)
	if err != nil {
		return fmt.Errorf("UpdateUserRole error: %w", err)
	}
	return nil
}

func GetAllUsers() ([]structs.User, error) {
	rows, err := Pool.Query(context.Background(),
		`SELECT char_id, name, role FROM users ORDER BY name;`)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers query error: %w", err)
	}
	defer rows.Close()

	var users []structs.User
	for rows.Next() {
		var u structs.User
		if err := rows.Scan(&u.CharID, &u.Name, &u.Role); err != nil {
			return nil, fmt.Errorf("GetAllUsers scan error: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Corp/Alliance nur persistieren (für spätere Anzeige/Checks)
func UpdateUserCorp(charID string, corpID int64, corpName, corpTicker string,
	allianceID *int64, allianceName, allianceTicker *string) error {

	q := `
	UPDATE users SET
	  corp_id = $2,
	  corp_name = $3,
	  corp_ticker = $4,
	  alliance_id = $5,
	  alliance_name = $6,
	  alliance_ticker = $7
	WHERE char_id = $1;`

	_, err := Pool.Exec(context.Background(), q,
		charID, corpID, corpName, corpTicker, allianceID, allianceName, allianceTicker)
	if err != nil {
		return fmt.Errorf("UpdateUserCorp: %w", err)
	}
	return nil
}
