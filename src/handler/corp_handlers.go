package handler

import (
	"net/http"

	db2 "speedliner-server/src/db"
)

func ListCorpsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	items, err := db2.SearchCorps(q, 100)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "DB error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}
