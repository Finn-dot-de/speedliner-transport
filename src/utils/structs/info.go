package structs

type RoleInfo struct {
	Name           string   `json:"name"`
	AppRoles       []string `json:"app_roles"`
	CorpRoles      []string `json:"corp_roles"`
	CorpID         *int64   `json:"corp_id,omitempty"`
	CorpName       *string  `json:"corp_name,omitempty"`
	CorpTicker     *string  `json:"corp_ticker,omitempty"`
	AllianceID     *int64   `json:"alliance_id,omitempty"`
	AllianceName   *string  `json:"alliance_name,omitempty"`
	AllianceTicker *string  `json:"alliance_ticker,omitempty"`
	OrgTag         string   `json:"org_tag"`
}
