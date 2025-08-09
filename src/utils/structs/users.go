package structs

// UserResponse godoc
// @Description Darstellung eines Users für die Admin-API.
type User struct {
	CharID string `json:"char_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type UpdateRoleReq struct {
	Role string `json:"role"`
}

var AllowedRoles = map[string]bool{
	"user":     true,
	"provider": true,
	"admin":    true,
}
