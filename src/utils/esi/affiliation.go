package esi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ua = "speedliner-server/1.0 (+affiliation)"

var httpClient = &http.Client{Timeout: 10 * time.Second}

type affiliationResp struct {
	CharacterID   int64  `json:"character_id"`
	CorporationID int64  `json:"corporation_id"`
	AllianceID    *int64 `json:"alliance_id,omitempty"`
}

type corpResp struct {
	CorporationID int64  `json:"corporation_id"`
	Name          string `json:"name"`
	Ticker        string `json:"ticker"`
	AllianceID    *int64 `json:"alliance_id,omitempty"`
}

type allianceResp struct {
	AllianceID int64  `json:"alliance_id"`
	Name       string `json:"name"`
	Ticker     string `json:"ticker"`
}

// einfache ETag-Caches (pro Prozess)
var etagCache = map[string]string{}
var bodyCache = map[string][]byte{}

// FetchCorpAndAlliance holt "jetzt"-Zugehörigkeit via Affiliation (Cache ~1h),
// und resolved Namen/Ticker. Kann sicher aus Callback aufgerufen werden.
func FetchCorpAndAlliance(characterID int) (corpID int64, corpName, corpTicker string, alliID *int64, alliName, alliTicker *string) {
	aff := fetchAffiliation(int64(characterID)) // robust: im Zweifel 0-Werte
	if aff == nil || aff.CorporationID == 0 {
		return 0, "", "", nil, nil, nil
	}

	// Corp-Details
	if c := fetchCorp(aff.CorporationID); c != nil {
		corpID = c.CorporationID
		corpName = c.Name
		corpTicker = c.Ticker
	}

	// Alliance-Details (wenn vorhanden)
	if aff.AllianceID != nil && *aff.AllianceID != 0 {
		if a := fetchAlliance(*aff.AllianceID); a != nil {
			aid := a.AllianceID
			alliID = &aid
			an, at := a.Name, a.Ticker
			alliName, alliTicker = &an, &at
		}
	}
	return
}

func fetchAffiliation(charID int64) *affiliationResp {
	url := "https://esi.evetech.net/v1/characters/affiliation/?datasource=tranquility"
	payload, _ := json.Marshal([]int64{charID})
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", ua)
	if etag := etagCache[url]; etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		b, _ := io.ReadAll(resp.Body)
		etagCache[url] = resp.Header.Get("ETag")
		bodyCache[url] = b
		var arr []affiliationResp
		if err := json.Unmarshal(b, &arr); err == nil && len(arr) > 0 {
			return &arr[0]
		}
		return nil
	case http.StatusNotModified:
		if b := bodyCache[url]; b != nil {
			var arr []affiliationResp
			if err := json.Unmarshal(b, &arr); err == nil && len(arr) > 0 {
				return &arr[0]
			}
		}
		return nil
	default:
		return nil
	}
}

func fetchCorp(id int64) *corpResp {
	url := fmt.Sprintf("https://esi.evetech.net/latest/corporations/%d/?datasource=tranquility", id)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", ua)
	if etag := etagCache[url]; etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		b, _ := io.ReadAll(resp.Body)
		etagCache[url] = resp.Header.Get("ETag")
		bodyCache[url] = b
		var out corpResp
		if err := json.Unmarshal(b, &out); err == nil {
			return &out
		}
	case http.StatusNotModified:
		if b := bodyCache[url]; b != nil {
			var out corpResp
			if err := json.Unmarshal(b, &out); err == nil {
				return &out
			}
		}
	}
	return nil
}

func fetchAlliance(id int64) *allianceResp {
	url := fmt.Sprintf("https://esi.evetech.net/latest/alliances/%d/?datasource=tranquility", id)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", ua)
	if etag := etagCache[url]; etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		b, _ := io.ReadAll(resp.Body)
		etagCache[url] = resp.Header.Get("ETag")
		bodyCache[url] = b
		var out allianceResp
		if err := json.Unmarshal(b, &out); err == nil {
			return &out
		}
	case http.StatusNotModified:
		if b := bodyCache[url]; b != nil {
			var out allianceResp
			if err := json.Unmarshal(b, &out); err == nil {
				return &out
			}
		}
	}
	return nil
}
