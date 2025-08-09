package handler

import (
	"encoding/json"
	"net/http"
	"speedliner-server/db"
	"speedliner-server/src/utils/structs"

	"github.com/go-chi/chi/v5"
)

// ListUsersHandler godoc
// @Summary      Alle Benutzer abrufen
// @Description  Gibt eine Liste aller Benutzer mit deren Rollen zurück (nur Admin).
// @Tags         Admin
// @Produce      json
// @Success      200 {array} structs.User
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB error"
// @Router       /app/users [get]
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := db.GetAllUsers()
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

// UpdateUserRoleHandler godoc
// @Summary      Rolle eines Benutzers ändern
// @Description  Setzt die Rolle eines Benutzers anhand der charID (nur Admin).
// @Tags         Admin
// @Accept       json
// @Produce      plain
// @Param        charID path string true "Character ID"
// @Param        role body structs.UpdateRoleReq true "Neue Rolle"
// @Success      204 {string} string "No Content"
// @Failure      400 {string} string "Invalid JSON or role"
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB error"
// @Router       /app/users/{charID}/role [put]
func UpdateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	charID := chi.URLParam(r, "charID")

	var req structs.UpdateRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if !structs.AllowedRoles[req.Role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	if err := db.UpdateUserRole(charID, req.Role); err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
