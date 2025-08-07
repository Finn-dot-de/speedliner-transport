package get

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"speedliner-server/db"
	"strconv"
	"time"

	"speedliner-server/src/utils/esiauth"
	"speedliner-server/src/utils/structs"

	"github.com/go-chi/chi/v5"
)

func DefineGetRoutes(r chi.Router) {
	r.Get("/ping", PingHandler)
	r.Get("/login", LoginHandler)
	r.Get("/callback", CallbackHandler)
	r.Get("/me", MeHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/routes", RoutesHandler)

}

// PingHandler godoc
// @Summary      Healthcheck
// @Description  Gibt "pong" zurück
// @Tags         System
// @Produce      plain
// @Success      200 {string} string "pong"
// @Router       /app/ping [get]
func PingHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

// LoginHandler godoc
// @Summary      Login redirect
// @Description  Leitet zum ESI-Login um
// @Tags         Auth
// @Produce      html
// @Success      302 {string} string "Redirect"
// @Router       /app/login [get]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	oauth := esiauth.GetOAuthConfig()
	url := oauth.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackHandler godoc
// @Summary      ESI Callback
// @Description  OAuth2 Callback für ESI, speichert Token und setzt Cookie
// @Tags         Auth
// @Produce      html
// @Param        code query string true "Authorization Code"
// @Success      302 {string} string "Redirect to home"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /app/callback [get]
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauth := esiauth.GetOAuthConfig()
	code := r.URL.Query().Get("code")

	token, err := oauth.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth.Client(context.Background(), token)
	resp, err := client.Get("https://esi.evetech.net/verify")
	if err != nil {
		http.Error(w, "Verify failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var verify structs.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verify); err != nil {
		http.Error(w, "Verify JSON error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	charID := fmt.Sprintf("%d", verify.CharacterID)
	esiauth.SaveToken(charID, token)

	http.SetCookie(w, &http.Cookie{
		Name:     "char_id",
		Value:    charID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	charID = strconv.Itoa(verify.CharacterID)
	charName := verify.CharacterName

	// In DB speichern (oder aktualisieren)
	if err := db.UpsertUser(charID, charName); err != nil {
		log.Printf("Failed to insert user: %v", err)
		// Optional: HTTP-Fehler zurück
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// MeHandler godoc
// @Summary      User Info
// @Description  Gibt Charakter-ID und Name zurück (aus Cookie)
// @Tags         Auth
// @Produce      json
// @Success      200 {object} structs.VerifyResponse
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /app/me [get]
func MeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("char_id")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	charID := cookie.Value
	token, ok := esiauth.LoadToken(charID)
	if !ok {
		http.Error(w, "No token for user", http.StatusUnauthorized)
		return
	}

	client := esiauth.GetOAuthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://esi.evetech.net/verify")
	if err != nil {
		http.Error(w, "Verify failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var verify structs.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verify); err != nil {
		http.Error(w, "Verify JSON error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(verify); err != nil {
		http.Error(w, "JSON encode error: "+err.Error(), http.StatusInternalServerError)
	}
}

// LogoutHandler godoc
// @Summary      Logout
// @Description  Löscht den Auth-Cookie und entfernt das Token aus dem Speicher
// @Tags         Auth
// @Produce      plain
// @Success      204 {string} string "No Content"
// @Router       /app/logout [get]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("char_id")
	if err == nil && cookie.Value != "" {
		esiauth.DeleteToken(cookie.Value)
	}

	// Cookie löschen
	http.SetCookie(w, &http.Cookie{
		Name:     "char_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0), // Abgelaufen
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

// RoutesHandler godoc
// @Summary      Get all active routes
// @Description  Gibt alle verfügbaren Transport-Routen zurück
// @Tags         Routes
// @Produce      json
// @Success      200 {array} structs.Route
// @Router       /app/routes [get]
func RoutesHandler(w http.ResponseWriter, r *http.Request) {
	routes, err := db.GetAllRoutes()
	if err != nil {
		http.Error(w, "Failed to fetch routes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(routes); err != nil {
		http.Error(w, "Failed to encode routes: "+err.Error(), http.StatusInternalServerError)
	}
}
