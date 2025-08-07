package db

import (
	"context"
)

type Route struct {
	ID                   string  `json:"id"`
	FromSystem           string  `json:"from"`
	ToSystem             string  `json:"to"`
	PricePerM3           float64 `json:"pricePerM3"`
	CollateralFeePercent float64 `json:"collateralFeePercent"`
	VolumeMax            float64 `json:"volumeMax"`
}

func GetAllRoutes() ([]Route, error) {
	rows, err := Pool.Query(context.Background(), `
		SELECT id, from_system, to_system, price_per_m3, collateral_fee_percent, volume_max
		FROM routes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Route
	for rows.Next() {
		var r Route
		err := rows.Scan(&r.ID, &r.FromSystem, &r.ToSystem, &r.PricePerM3, &r.CollateralFeePercent, &r.VolumeMax)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}
