package db

import (
	"context"
	"speedliner-server/src/utils/structs"

	"github.com/jackc/pgx/v5"
)

func InsertRoute(r structs.Route) error {
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

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO routes (id, from_system, to_system, price_per_m3, no_collateral, visibility)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		RETURNING id`,
		r.From, r.To, r.PricePerM3, r.NoCollateral, r.Visibility,
	).Scan(&id)
	if err != nil {
		return err
	}

	if r.Visibility == "whitelist" && len(r.AllowedCorpIDs) > 0 {
		b := &pgx.Batch{}
		for _, cid := range r.AllowedCorpIDs {
			b.Queue(`INSERT INTO route_visibility(route_id, corp_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, cid)
		}
		br := tx.SendBatch(ctx, b)
		if err = br.Close(); err != nil {
			return err
		}
	}
	return nil
}

func UpdateRoute(r structs.Route) error {
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

	_, err = tx.Exec(ctx, `
		UPDATE routes
		SET from_system=$1, to_system=$2, price_per_m3=$3, no_collateral=$4, visibility=$5
		WHERE id=$6`,
		r.From, r.To, r.PricePerM3, r.NoCollateral, r.Visibility, r.ID)
	if err != nil {
		return err
	}

	// Whitelist neu setzen
	_, err = tx.Exec(ctx, `DELETE FROM route_visibility WHERE route_id=$1`, r.ID)
	if err != nil {
		return err
	}
	if r.Visibility == "whitelist" && len(r.AllowedCorpIDs) > 0 {
		b := &pgx.Batch{}
		for _, cid := range r.AllowedCorpIDs {
			b.Queue(`INSERT INTO route_visibility(route_id, corp_id) VALUES ($1,$2)`, r.ID, cid)
		}
		br := tx.SendBatch(ctx, b)
		if err = br.Close(); err != nil {
			return err
		}
	}
	return nil
}

func GetAllRoutesForUser(charID *int64) ([]structs.Route, error) {
	ctx := context.Background()

	var rows pgx.Rows
	var err error

	if charID == nil {
		rows, err = Pool.Query(ctx, `
			SELECT r.id, r.from_system, r.to_system, r.price_per_m3, r.no_collateral, r.visibility,
			       COALESCE(rv.allowed, ARRAY[]::bigint[]) AS allowed
			FROM routes r
			LEFT JOIN LATERAL (
				SELECT ARRAY_AGG(v.corp_id ORDER BY v.corp_id) AS allowed
				FROM route_visibility v
				WHERE v.route_id = r.id
			) rv ON TRUE
			WHERE r.visibility = 'all'
			ORDER BY r.from_system, r.to_system`)
	} else {
		rows, err = Pool.Query(ctx, `
			SELECT r.id, r.from_system, r.to_system, r.price_per_m3, r.no_collateral, r.visibility,
			       COALESCE(rv.allowed, ARRAY[]::bigint[]) AS allowed
			FROM routes r
			LEFT JOIN LATERAL (
				SELECT ARRAY_AGG(v.corp_id ORDER BY v.corp_id) AS allowed
				FROM route_visibility v
				WHERE v.route_id = r.id
			) rv ON TRUE
			WHERE r.visibility = 'all'
			   OR (r.visibility = 'whitelist' AND EXISTS (
					SELECT 1
					FROM route_visibility v
					WHERE v.route_id = r.id
					  AND v.corp_id = (SELECT corp_id FROM users WHERE char_id = $1)
			   ))
			ORDER BY r.from_system, r.to_system`, *charID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []structs.Route
	for rows.Next() {
		var rr structs.Route
		if err := rows.Scan(&rr.ID, &rr.From, &rr.To, &rr.PricePerM3, &rr.NoCollateral, &rr.Visibility, &rr.AllowedCorpIDs); err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

// DeleteRoute löscht eine Route anhand ihrer ID
func DeleteRoute(id string) error {
	_, err := Pool.Exec(context.Background(), `
		DELETE FROM routes WHERE id = $1
	`, id)
	return err
}
