package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	db2 "speedliner-server/src/db"
	"speedliner-server/src/middleware"
	"speedliner-server/src/utils/esi"
	"strconv"
	"strings"
	"time"

	"speedliner-server/src/utils/esiauth"
	"speedliner-server/src/utils/structs"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

func DefineApiRoutes(r chi.Router) {
	r.Get("/ping", PingHandler)
	r.Get("/login", LoginHandler)
	r.Get("/callback", CallbackHandler)
	r.Get("/me", MeHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/routes", RoutesHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Post("/routes", CreateRouteHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Put("/routes/{id}", UpdateRouteHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Delete("/routes/{id}", DeleteRouteHandler)
	r.Get("/role", GetUserRoleHandler)
	r.With(middleware.RoleMiddleware("admin")).Get("/users", ListUsersHandler)
	r.With(middleware.RoleMiddleware("admin")).Put("/users/{charID}/role", UpdateUserRoleHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Get("/corps", ListCorpsHandler)
	r.Post("/mail", SendMailHandler)
	r.Post("/express/mail", SendExpressMailFromServiceHandler)
	r.With(middleware.RoleMiddleware("admin")).Get("/express/token-status", ExpressTokenStatusHandler)

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
	defer resp.Body.Close()

	var verify structs.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verify); err != nil {
		http.Error(w, "Verify JSON error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Token speichern & Cookie setzen
	charIDStr := strconv.Itoa(verify.CharacterID)
	esiauth.SaveToken(charIDStr, token)
	http.SetCookie(w, &http.Cookie{
		Name:     "char",
		Value:    charIDStr,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(48 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	// DB: alles als int64
	charID := int64(verify.CharacterID)

	// User upserten
	if err := db2.UpsertUser(charID, verify.CharacterName); err != nil {
		log.Printf("UpsertUser: %v", err)
	}

	// Corp/Alliance holen
	corpID, corpName, corpTicker, alliID, alliName, alliTicker :=
		esi.FetchCorpAndAlliance(verify.CharacterID)

	// Alliance (optional)
	var alliPtr *int64
	if alliID != nil && *alliID != 0 {
		// alliName/alliTicker können nil sein -> leere Strings
		var aName, aTick string
		if alliName != nil {
			aName = *alliName
		}
		if alliTicker != nil {
			aTick = *alliTicker
		}
		if err := db2.UpsertAlliance(*alliID, aName, aTick); err != nil {
			log.Printf("UpsertAlliance: %v", err)
		}
		alliPtr = alliID
	}

	// Corp (wenn vorhanden) + beim User setzen
	if corpID != 0 {
		if err := db2.UpsertCorp(corpID, corpName, corpTicker, alliPtr); err != nil {
			log.Printf("UpsertCorp: %v", err)
		}
		if err := db2.UpdateUserCorp(charID, corpID); err != nil {
			log.Printf("UpdateUserCorp: %v", err)
		}
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
	cookie, err := r.Cookie("char")
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
	cookie, err := r.Cookie("char")
	if err == nil && cookie.Value != "" {
		esiauth.DeleteToken()
	}

	// Cookie löschen
	http.SetCookie(w, &http.Cookie{
		Name:     "char",
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
// src/handler/handler.go (RoutesHandler)
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
		http.Error(w, "Failed to fetch routes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(routes)
}

// CreateRouteHandler godoc
// @Summary      Neue Route anlegen
// @Description  Erstellt eine neue Transport-Route
// @Tags         Routes
// @Accept       json
// @Produce      json
// @Param        route body structs.Route true "Neue Route"
// @Success      201 {object} structs.Route
// @Failure      400 {string} string "Invalid JSON"
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB Insert error"
// @Router       /app/routes [post]
func CreateRouteHandler(w http.ResponseWriter, r *http.Request) {
	var route structs.Route

	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := db2.InsertRoute(route); err != nil {
		http.Error(w, "DB Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(route)
}

// UpdateRouteHandler godoc
// @Summary      Route aktualisieren
// @Description  Aktualisiert eine bestehende Transport-Route anhand der ID
// @Tags         Routes
// @Accept       json
// @Produce      json
// @Param        id path string true "Route ID"
// @Param        route body structs.Route true "Route-Daten"
// @Success      200 {object} structs.Route
// @Failure      400 {string} string "Invalid JSON"
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB Update error"
// @Router       /app/routes/{id} [put]
func UpdateRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var route structs.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	route.ID = id
	if err := db2.UpdateRoute(route); err != nil {
		http.Error(w, "DB Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(route)
}

// DeleteRouteHandler godoc
// @Summary      Route löschen
// @Description  Löscht eine Transport-Route anhand der ID
// @Tags         Routes
// @Produce      plain
// @Param        id path string true "Route ID"
// @Success      204 {string} string "Deleted"
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB Delete error"
// @Router       /app/routes/{id} [delete]
func DeleteRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := db2.DeleteRoute(id); err != nil {
		http.Error(w, "DB Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetUserRoleHandler godoc
// @Summary      Benutzerrolle abrufen
// @Description  Gibt die Rolle des eingeloggten Benutzers zurück
// @Tags         Auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      401 {string} string "Unauthorized"
// @Failure      403 {string} string "Forbidden"
// @Failure      500 {string} string "DB error"
// @Router       /app/role [get]
func GetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("char")
	if err != nil || c.Value == "" {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}
	charID, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		http.Error(w, "Bad char id", http.StatusUnauthorized)
		return
	}

	role, err := db2.GetUserRoles(charID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"role": role})
}

func ListCorpsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	items, err := db2.SearchCorps(q, 100)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// SendMailHandler godoc
// @Summary      EVE-Mail senden (ohne CSPA)
// @Description  Sendet eine EVE-Mail über ESI. Empfänger nutzt kein CSPA, daher wird `approved_cost` immer 0 gesendet.
// @Tags         Mail
// @Accept       json
// @Produce      json
// @Param        mail body structs.SendMailRequest true "Mail-Daten" example({"subject":"Test","body":"Hello World","recipients":[{"id":2118431553,"type":"character"}]})
// @Success      201 {object} structs.MailIDResponse
// @Failure      400 {object} structs.ErrorResponse
// @Failure      401 {object} structs.ErrorResponse
// @Failure      502 {object} structs.ErrorResponse
// @Router       /app/mail [post]
func SendMailHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("char")
	if err != nil || c.Value == "" {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	token, ok := esiauth.LoadToken(c.Value)
	if !ok {
		http.Error(w, "No token for user", http.StatusUnauthorized)
		return
	}
	httpClient := esiauth.GetOAuthConfig().Client(context.Background(), token)

	// Request einlesen (du hast die Structs ja schon)
	var req structs.SendMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Subject == "" || req.Body == "" || len(req.Recipients) == 0 {
		http.Error(w, "subject, body, recipients required", http.StatusBadRequest)
		return
	}

	// Empfänger in ESI-Format bringen + minimal prüfen
	allowed := map[string]bool{"character": true, "corporation": true, "alliance": true, "mailing_list": true}
	recipients := make([]map[string]interface{}, 0, len(req.Recipients))
	for _, rcpt := range req.Recipients {
		if rcpt.ID <= 0 || !allowed[rcpt.Type] {
			http.Error(w, "invalid recipient entry", http.StatusBadRequest)
			return
		}
		recipients = append(recipients, map[string]interface{}{
			"recipient_id":   rcpt.ID,
			"recipient_type": rcpt.Type,
		})
	}

	// ESI-Aufruf (approved_cost = 0 reicht; Default bei ESI ist ebenfalls 0)
	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%s/mail/?datasource=tranquility", c.Value)
	payload := map[string]interface{}{
		"approved_cost": 0,
		"subject":       req.Subject,
		"body":          req.Body,
		"recipients":    recipients,
	}
	body, _ := json.Marshal(payload)

	reqESI, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	reqESI.Header.Set("Content-Type", "application/json")
	reqESI.Header.Set("User-Agent", "speedliner-server/1.0 (mail)")

	resp, err := httpClient.Do(reqESI)
	if err != nil {
		http.Error(w, "ESI error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("ESI send failed: %s: %s", resp.Status, string(b)), http.StatusBadGateway)
		return
	}

	raw, _ := io.ReadAll(resp.Body)
	mailID, _ := strconv.Atoi(strings.TrimSpace(string(raw)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int{"mail_id": mailID})
}

// SendExpressMailFromServiceHandler godoc
// @Summary     EVE-Mail für EXPRESS senden (Service-Char -> Ziel-Corp)
// @Description Verwendet einen fest konfigurierten Service-Char (ENV), sendet Mail an eine Ziel-Corp (ENV).
// @Tags        Mail
// @Accept      json
// @Produce     json
// @Param       body body structs.ExpressMailRequest true "Express Daten"
// @Success     201 {object} map[string]int "mail_id"
// @Failure     400 {string} string
// @Failure     401 {string} string
// @Failure     500 {string} string
// @Failure     502 {string} string
// @Router      /app/express/mail [post]
func SendExpressMailFromServiceHandler(w http.ResponseWriter, r *http.Request) {
	senderCharID := strings.TrimSpace(os.Getenv("EXPRESS_SENDER_CHAR_ID"))
	targetKind := strings.TrimSpace(os.Getenv("EXPRESS_TARGET_TYPE"))
	targetIDStr := strings.TrimSpace(os.Getenv("EXPRESS_TARGET_CORP_ID"))

	if senderCharID == "" || targetIDStr == "" {
		http.Error(w, "missing EXPRESS_SENDER_CHAR_ID or EXPRESS_TARGET_CORP_ID", http.StatusInternalServerError)
		return
	}
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil || targetID <= 0 {
		http.Error(w, "bad EXPRESS_TARGET_CORP_ID", http.StatusInternalServerError)
		return
	}

	var req structs.ExpressMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !req.Express || strings.TrimSpace(req.Route) == "" || req.RewardISK <= 0 || req.VolumeM3 <= 0 {
		http.Error(w, "missing required express fields", http.StatusBadRequest)
		return
	}

	// Empfänger gegen ESI prüfen → vermeidet 400 "Bad corporation or allianceID"
	if ok, status, verr := validateRecipient(targetKind, targetID); verr != nil {
		http.Error(w, "validate recipient error: "+verr.Error(), http.StatusBadGateway)
		return
	} else if !ok {
		http.Error(w, fmt.Sprintf("invalid %s id %d (ESI %s)", targetKind, targetID, status), http.StatusBadRequest)
		return
	}

	tok, ok := esiauth.LoadToken(senderCharID)
	if !ok {
		http.Error(w, "service token missing (login service char once)", http.StatusUnauthorized)
		return
	}

	cfg := esiauth.GetOAuthConfig()
	ctx := context.Background()
	baseTS := cfg.TokenSource(ctx, tok)
	ts := esiauth.NewSavingTokenSource(senderCharID, baseTS)
	httpClient := oauth2.NewClient(ctx, ts)

	subject := fmt.Sprintf("EXPRESS: %s — %s ISK", req.Route, formatISK(req.RewardISK))
	body := buildExpressMailBody(req)

	payload := map[string]interface{}{
		"approved_cost": 0,
		"subject":       subject,
		"body":          body,
		"recipients": []map[string]interface{}{
			{
				"recipient_id":   targetID,
				"recipient_type": targetKind, // <— wichtig
			},
		},
	}
	bts, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%s/mail/?datasource=tranquility", senderCharID)
	reqESI, _ := http.NewRequest("POST", url, bytes.NewReader(bts))
	reqESI.Header.Set("Content-Type", "application/json")
	reqESI.Header.Set("User-Agent", "speedliner-server/1.0 (express-mail)")

	resp, err := httpClient.Do(reqESI)
	if err != nil {
		http.Error(w, "ESI error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("ESI send failed: %s: %s", resp.Status, string(raw)), http.StatusBadGateway)
		return
	}

	raw, _ := io.ReadAll(resp.Body)
	mailID, _ := strconv.Atoi(strings.TrimSpace(string(raw)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int{"mail_id": mailID})
}

func ExpressTokenStatusHandler(w http.ResponseWriter, r *http.Request) {
	senderCharID := strings.TrimSpace(os.Getenv("EXPRESS_SENDER_CHAR_ID"))
	if senderCharID == "" {
		http.Error(w, "missing ENV EXPRESS_SENDER_CHAR_ID", http.StatusInternalServerError)
		return
	}

	// 1) Token aus Store laden (existiert?)
	tok, ok := esiauth.LoadToken(senderCharID)

	// 2) updated_at aus DB lesen
	var updatedAt *time.Time
	if db2.Pool != nil {
		var ua time.Time
		err := db2.Pool.QueryRow(context.Background(),
			`SELECT updated_at FROM oauth_tokens WHERE char_id=$1`, senderCharID,
		).Scan(&ua)
		if err == nil {
			updatedAt = &ua
		} else if !errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "db query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 3) Expiry-Infos aus dem Access-Token (falls vorhanden)
	var exp *time.Time
	var minsLeft *int64
	var expired *bool
	if tok != nil {
		t := tok.Expiry
		exp = &t
		ml := int64(time.Until(t).Minutes())
		minsLeft = &ml
		ex := time.Now().After(t)
		expired = &ex
	}

	// 4) Antwort bauen
	resp := map[string]interface{}{
		"sender_char_id":       senderCharID,
		"has_token":            ok && tok != nil,
		"updated_at":           nil,
		"access_token_expiry":  nil,
		"minutes_until_expiry": nil,
		"access_token_expired": nil,
	}
	if updatedAt != nil {
		resp["updated_at"] = updatedAt.UTC().Format(time.RFC3339)
	}
	if exp != nil {
		resp["access_token_expiry"] = exp.UTC().Format(time.RFC3339)
	}
	if minsLeft != nil {
		resp["minutes_until_expiry"] = *minsLeft
	}
	if expired != nil {
		resp["access_token_expired"] = *expired
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func validateRecipient(kind string, id int64) (bool, string, error) {
	var url string
	switch kind {
	case "corporation":
		url = fmt.Sprintf("https://esi.evetech.net/latest/corporations/%d/?datasource=tranquility", id)
	case "alliance":
		url = fmt.Sprintf("https://esi.evetech.net/latest/alliances/%d/?datasource=tranquility", id)
	default:
		return false, "", fmt.Errorf("unsupported target kind: %s", kind)
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "speedliner-server/1.0 (validate)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, resp.Status, nil
	}
	return true, "OK", nil
}
