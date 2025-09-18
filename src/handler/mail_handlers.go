package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	db2 "speedliner-server/src/db"
	"speedliner-server/src/utils/esiauth"
	"speedliner-server/src/utils/structs"

	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

// /mail – sendet als eingeloggter User
func SendMailHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("char")
	if err != nil || c.Value == "" {
		jsonError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	token, ok := esiauth.LoadToken(c.Value)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "No token for user")
		return
	}
	httpClient := esiauth.GetOAuthConfig().Client(context.Background(), token)

	var req structs.SendMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if req.Subject == "" || req.Body == "" || len(req.Recipients) == 0 {
		jsonError(w, http.StatusBadRequest, "subject, body, recipients required")
		return
	}

	allowed := map[string]bool{"character": true, "corporation": true, "alliance": true, "mailing_list": true}
	recipients := make([]map[string]interface{}, 0, len(req.Recipients))
	for _, rcpt := range req.Recipients {
		if rcpt.ID <= 0 || !allowed[rcpt.Type] {
			jsonError(w, http.StatusBadRequest, "invalid recipient entry")
			return
		}
		recipients = append(recipients, map[string]interface{}{
			"recipient_id":   rcpt.ID,
			"recipient_type": rcpt.Type,
		})
	}

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
		jsonError(w, http.StatusBadGateway, "ESI error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		jsonError(w, http.StatusBadGateway, fmt.Sprintf("ESI send failed: %s: %s", resp.Status, string(b)))
		return
	}

	raw, _ := io.ReadAll(resp.Body)
	mailID, _ := strconv.Atoi(strings.TrimSpace(string(raw)))

	writeJSON(w, http.StatusCreated, map[string]int{"mail_id": mailID})
}

// EXPRESS: sendet als Service-Char an Ziel-Corp/Alliance
func SendExpressMailFromServiceHandler(w http.ResponseWriter, r *http.Request) {
	senderCharID := strings.TrimSpace(os.Getenv("EXPRESS_SENDER_CHAR_ID"))
	targetKind := strings.TrimSpace(os.Getenv("EXPRESS_TARGET_TYPE"))
	targetIDStr := strings.TrimSpace(os.Getenv("EXPRESS_TARGET_CORP_ID"))

	if senderCharID == "" || targetIDStr == "" {
		jsonError(w, http.StatusInternalServerError, "missing EXPRESS_SENDER_CHAR_ID or EXPRESS_TARGET_CORP_ID")
		return
	}
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil || targetID <= 0 {
		jsonError(w, http.StatusInternalServerError, "bad EXPRESS_TARGET_CORP_ID")
		return
	}

	var req structs.ExpressMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if !req.Express || strings.TrimSpace(req.Route) == "" || req.RewardISK <= 0 || req.VolumeM3 <= 0 {
		jsonError(w, http.StatusBadRequest, "missing required express fields")
		return
	}

	// Empfänger prüfen
	if ok, status, verr := validateRecipient(targetKind, targetID); verr != nil {
		jsonError(w, http.StatusBadGateway, "validate recipient error: "+verr.Error())
		return
	} else if !ok {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("invalid %s id %d (ESI %s)", targetKind, targetID, status))
		return
	}

	tok, ok := esiauth.LoadToken(senderCharID)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "service token missing (login service char once)")
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
			{"recipient_id": targetID, "recipient_type": targetKind},
		},
	}
	bts, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%s/mail/?datasource=tranquility", senderCharID)
	reqESI, _ := http.NewRequest("POST", url, bytes.NewReader(bts))
	reqESI.Header.Set("Content-Type", "application/json")
	reqESI.Header.Set("User-Agent", "speedliner-server/1.0 (express-mail)")

	resp, err := httpClient.Do(reqESI)
	if err != nil {
		jsonError(w, http.StatusBadGateway, "ESI error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		jsonError(w, http.StatusBadGateway, fmt.Sprintf("ESI send failed: %s: %s", resp.Status, string(raw)))
		return
	}

	raw, _ := io.ReadAll(resp.Body)
	mailID, _ := strconv.Atoi(strings.TrimSpace(string(raw)))

	writeJSON(w, http.StatusCreated, map[string]int{"mail_id": mailID})
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

// kleine Helfer aus deinem Original
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
