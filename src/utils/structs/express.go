package structs

// ExpressMailRequest wird vom Frontend geschickt
type ExpressMailRequest struct {
	Route     string `json:"route"`           // "Amarr ↔ K-6K16"
	RewardISK int64  `json:"reward_isk"`      // z.B. 826500000
	VolumeM3  int64  `json:"volume_m3"`       // z.B. 165000
	CollatISK int64  `json:"collateral_isk"`  // 0..20B
	Express   bool   `json:"express"`         // true
	Notes     string `json:"notes,omitempty"` // optional
	// optional: Wer hat ausgelöst?
	CustomerCharID   int64  `json:"customer_char_id,omitempty"`
	CustomerCharName string `json:"customer_char_name,omitempty"`
}
