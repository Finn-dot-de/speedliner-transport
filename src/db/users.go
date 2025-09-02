package db

import (
	"context"
	"fmt"
	"speedliner-server/src/utils/structs"
)

func UpsertUser(charID int64, name string) error {
	_, err := Pool.Exec(context.Background(), `
		INSERT INTO users (char_id, name)
		VALUES ($1, $2)
		ON CONFLICT (char_id) DO UPDATE SET name = EXCLUDED.name;`,
		charID, name)
	return err
}

func GetUserRoles(charID int64) (string, error) {
	row := Pool.QueryRow(context.Background(),
		`SELECT role FROM users WHERE char_id = $1;`, charID)
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

func UpsertAlliance(id int64, name, ticker string) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO alliances (alliance_id,name,ticker)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (alliance_id) DO UPDATE
		 SET name=EXCLUDED.name, ticker=EXCLUDED.ticker`, id, name, ticker)
	return err
}

func UpsertCorp(id int64, name, ticker string, allianceID *int64) error {
	_, err := Pool.Exec(context.Background(),
		`INSERT INTO corps (corp_id,name,ticker,alliance_id)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (corp_id) DO UPDATE
		 SET name=EXCLUDED.name, ticker=EXCLUDED.ticker, alliance_id=EXCLUDED.alliance_id`,
		id, name, ticker, allianceID)
	return err
}

func UpdateUserCorp(charID, corpID int64) error {
	_, err := Pool.Exec(context.Background(),
		`UPDATE users SET corp_id=$2 WHERE char_id=$1`, charID, corpID)
	return err
}
