package handler

import (
	"encoding/json"
	"net/http"
)

// einheitliche JSON-Antworten
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}
