package handler

import (
	"encoding/json"
	"net/http"
	"speedliner-server/src/utils/structs"
	"strconv"

	db2 "speedliner-server/src/db"

	"github.com/go-chi/chi/v5"
)

func RoutesHandler(w http.ResponseWriter, r *http.Request) {
	var charID *int64
	var role string

	if c, err := r.Cookie("char"); err == nil && c.Value != "" {
		if v, err2 := strconv.ParseInt(c.Value, 10, 64); err2 == nil {
			charID = &v
			if rr, err3 := db2.GetUserRoles(v); err3 == nil {
				role = rr
			}
		}
	}

	routes, err := db2.GetAllRoutesForUser(charID, role)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to fetch routes: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, routes)
}

func CreateRouteHandler(w http.ResponseWriter, r *http.Request) {
	var route structs.Route // oder structs.Route – je nach deiner Definition
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if err := db2.InsertRoute(route); err != nil {
		jsonError(w, http.StatusInternalServerError, "DB Insert error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, route)
}

func UpdateRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var route structs.Route // oder structs.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	route.ID = id
	if err := db2.UpdateRoute(route); err != nil {
		jsonError(w, http.StatusInternalServerError, "DB Update error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, route)
}

func DeleteRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := db2.DeleteRoute(id); err != nil {
		jsonError(w, http.StatusInternalServerError, "DB Delete error: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
