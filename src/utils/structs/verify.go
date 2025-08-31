package structs

type VerifyResponse struct {
	CharacterID   int    `json:"CharacterID" example:"12345678"`
	CharacterName string `json:"CharacterName" example:"Pilot McFly"`
}
