package db

import "context"

type CorpOption struct {
	CorpID int64  `json:"corpId"`
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
}

func SearchCorps(q string, limit int) ([]CorpOption, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := Pool.Query(context.Background(), `
		SELECT corp_id, COALESCE(ticker,''), name
		FROM corps
		WHERE $1 = '' OR name ILIKE '%'||$1||'%' OR ticker ILIKE '%'||$1||'%'
		ORDER BY ticker NULLS LAST, name
		LIMIT $2`, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []CorpOption
	for rows.Next() {
		var it CorpOption
		if err := rows.Scan(&it.CorpID, &it.Ticker, &it.Name); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, rows.Err()
}
