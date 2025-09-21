package db

import (
	"context"
	"fmt"
	"speedliner-server/src/utils/structs"
)

func InsertRoute(r structs.Route) (err error) {
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

	// Fallback auf Default 50 Mio, wenn nichts/0 geliefert
	minPrice := r.MinPrice
	if minPrice == 0 || minPrice < 0 {
		minPrice = 50_000_000
	}
	if minPrice < 0 {
		return fmt.Errorf("min_price must be >= 0")
	}

	row := tx.QueryRow(ctx, `
        INSERT INTO routes (from_system, to_system, price_per_m3, no_collateral, visibility, min_price)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING id`,
		r.From, r.To, r.PricePerM3, r.NoCollateral, r.Visibility, minPrice,
	)
	if err = row.Scan(&r.ID); err != nil {
		return err
	}

	if r.Visibility == "whitelist" && len(r.AllowedCorps) > 0 {
		for _, cid := range r.AllowedCorps {
			if _, err = tx.Exec(ctx, `
                INSERT INTO route_visibility (route_id, corp_id)
                VALUES ($1,$2) ON CONFLICT DO NOTHING`, r.ID, cid); err != nil {
				return err
			}
		}
	}
	return nil
}

func UpdateRoute(r structs.Route) (err error) {
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

	if r.MinPrice == 0 || r.MinPrice < 0 {
		r.MinPrice = 50_000_000
	}

	if _, err = tx.Exec(ctx, `
        UPDATE routes
           SET from_system=$2,
               to_system=$3,
               price_per_m3=$4,
               no_collateral=$5,
               visibility=$6,
               min_price=$7
         WHERE id = $1`,
		r.ID, r.From, r.To, r.PricePerM3, r.NoCollateral, r.Visibility, r.MinPrice); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `DELETE FROM route_visibility WHERE route_id=$1`, r.ID); err != nil {
		return err
	}
	if r.Visibility == "whitelist" && len(r.AllowedCorps) > 0 {
		for _, cid := range r.AllowedCorps {
			if _, err = tx.Exec(ctx, `
                INSERT INTO route_visibility (route_id, corp_id)
                VALUES ($1,$2) ON CONFLICT DO NOTHING`, r.ID, cid); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetAllRoutesForUser(charID *int64, role string) ([]structs.Route, error) {
	ctx := context.Background()

	// Provider/Admin? -> ungefiltert
	if role == "admin" || role == "provider" {
		rows, err := Pool.Query(ctx, `
             SELECT 
                 r.id, 
                 r.from_system, 
                 r.to_system, 
                 r.price_per_m3, 
                 r.no_collateral, 
                 r.visibility, 
                 r.min_price
				FROM 
				    routes r
		  		ORDER BY 
		  		    r.from_system, 
		  		    r.to_system`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var list []structs.Route
		for rows.Next() {
			var it structs.Route
			if err := rows.Scan(&it.ID, &it.From, &it.To, &it.PricePerM3, &it.NoCollateral, &it.Visibility, &it.MinPrice); err != nil {
				return nil, err
			}
			list = append(list, it)
		}
		return list, rows.Err()
	}

	// normale User (eingeloggt oder anonym)
	if charID == nil {
		rows, err := Pool.Query(ctx, `
             SELECT 
                 r.id, 
                 r.from_system, 
                 r.to_system, 
                 r.price_per_m3, 
                 r.no_collateral, 
                 r.visibility, 
                 r.min_price
			 FROM 
			     routes r
			 WHERE 
			     r.visibility='all'
			 ORDER BY 
			     r.from_system, 
			     r.to_system`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var list []structs.Route
		for rows.Next() {
			var it structs.Route
			if err := rows.Scan(&it.ID, &it.From, &it.To, &it.PricePerM3, &it.NoCollateral, &it.Visibility, &it.MinPrice); err != nil {
				return nil, err
			}
			list = append(list, it)
		}
		return list, rows.Err()
	}

	rows, err := Pool.Query(ctx, `
        SELECT DISTINCT 
            r.id, 
            r.from_system, 
            r.to_system, 
            r.price_per_m3, 
            r.no_collateral, 
            r.visibility, 
            r.min_price
		FROM 
		    routes r
		JOIN 
		        users u ON u.char_id = $1
		WHERE 
		    r.visibility='all'
			OR (r.visibility='whitelist' AND EXISTS (
				  SELECT 1 
				  FROM 
				      route_visibility rv
				   WHERE 
				       rv.route_id=r.id 
				     AND 
				       rv.corp_id=u.corp_id
				))
	  	ORDER BY 
	  	    r.from_system, 
	  	    r.to_system`, *charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []structs.Route
	for rows.Next() {
		var it structs.Route
		if err := rows.Scan(&it.ID, &it.From, &it.To, &it.PricePerM3, &it.NoCollateral, &it.Visibility, &it.MinPrice); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, rows.Err()
}

func DeleteRoute(id string) error {
	_, err := Pool.Exec(context.Background(), `DELETE FROM routes WHERE id = $1`, id)
	return err
}
