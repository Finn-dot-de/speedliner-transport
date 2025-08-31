package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type corpInfo struct {
	Name       string `json:"name"`
	Ticker     string `json:"ticker"`
	AllianceID *int64 `json:"alliance_id,omitempty"`
}
type allianceInfo struct {
	Name   string `json:"name"`
	Ticker string `json:"ticker"`
}

// Beim Login callen; schreibt NICHT selbst in DB, nur Daten liefern
func FetchCorpAndAlliance(charID int) (corpID int64, corpName, corpTicker string,
	alliID *int64, alliName, alliTicker *string) {

	hc := &http.Client{Timeout: 8 * time.Second}

	// 1) Character -> corporation_id
	var c struct {
		CorporationID int64 `json:"corporation_id"`
	}
	if r, err := hc.Get(fmt.Sprintf("https://esi.evetech.net/v5/characters/%d/", charID)); err == nil {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&c)
	}
	corpID = c.CorporationID

	// 2) Corporation -> name, ticker, alliance_id
	var ci corpInfo
	if r, err := hc.Get(fmt.Sprintf("https://esi.evetech.net/v5/corporations/%d/", corpID)); err == nil {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&ci)
	}
	corpName, corpTicker, alliID = ci.Name, ci.Ticker, ci.AllianceID

	// 3) Alliance (optional)
	if alliID != nil {
		var ai allianceInfo
		if r, err := hc.Get(fmt.Sprintf("https://esi.evetech.net/v4/alliances/%d/", *alliID)); err == nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&ai); err == nil {
				alliName = &ai.Name
				alliTicker = &ai.Ticker
			}
		}
	}
	return
}
