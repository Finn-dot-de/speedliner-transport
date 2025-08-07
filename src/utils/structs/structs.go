package structs

type EsiSearchResponse struct {
	InventoryType []int `json:"inventory_type"`
}

type Item struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}
