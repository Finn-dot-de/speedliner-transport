package db

import (
	"context"
	"speedliner-server/src/utils/structs"
)

// InsertRoute fügt eine neue Route in die DB ein
func InsertRoute(route structs.Route) error {
	_, err := Pool.Exec(context.Background(), `
		INSERT INTO routes (id, from_system, to_system, price_per_m3, collateral_fee_percent, volume_max)
		VALUES (gen_random_uuid(), $1, $2, $3, 0.03, 337000)
	`, route.From, route.To, route.PricePerM3, route.CollateralFeePercent, route.VolumeMax)
	return err
}

// UpdateRoute aktualisiert eine bestehende Route anhand ihrer ID
func UpdateRoute(route structs.Route) error {
	_, err := Pool.Exec(context.Background(), `
		UPDATE routes
		SET from_system = $1,
		    to_system = $2,
		    price_per_m3 = $3
		WHERE id = $6
	`, route.From, route.To, route.PricePerM3, route.CollateralFeePercent, route.VolumeMax, route.ID)
	return err
}

// DeleteRoute löscht eine Route anhand ihrer ID
func DeleteRoute(id string) error {
	_, err := Pool.Exec(context.Background(), `
		DELETE FROM routes WHERE id = $1
	`, id)
	return err
}

// GetAllRoutes gibt alle Routen aus der DB zurück
func GetAllRoutes() ([]structs.Route, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT id, from_system, to_system, price_per_m3, collateral_fee_percent, volume_max
		FROM routes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []structs.Route
	for rows.Next() {
		var r structs.Route
		if err := rows.Scan(&r.ID, &r.From, &r.To, &r.PricePerM3, &r.CollateralFeePercent, &r.VolumeMax); err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}

	return routes, nil
}
