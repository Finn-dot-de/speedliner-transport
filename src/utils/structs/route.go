package structs

type Route struct {
	ID           string  `json:"id"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	PricePerM3   float64 `json:"pricePerM3"`
	NoCollateral bool    `json:"noCollateral"`
	Visibility   string  `json:"visibility"` // "all" | "whitelist"
	AllowedCorps []int64 `json:"allowedCorps,omitempty"`
}
