package structs

type Route struct {
	ID           string `json:"id"`
	From         string `json:"from"`
	To           string `json:"to"`
	PricePerM3   int    `json:"pricePerM3"`
	NoCollateral bool   `json:"noCollateral"`
}
