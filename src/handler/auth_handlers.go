package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	db2 "speedliner-server/src/db"
	"speedliner-server/src/utils/esi"
	"speedliner-server/src/utils/esiauth"
	"speedliner-server/src/utils/structs"
)

// Health
func PingHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

// Login redirect
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	oauth := esiauth.GetOAuthConfig()
	url := oauth.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusFound)
}

// OAuth Callback —> speichert Token, setzt Cookie, resolved Corp/Alliance via Affiliation
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauth := esiauth.GetOAuthConfig()
	code := r.URL.Query().Get("code")

	token, err := oauth.Exchange(context.Background(), code)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Token exchange failed: "+err.Error())
		return
	}

	client := oauth.Client(context.Background(), token)
	resp, err := client.Get("https://esi.evetech.net/verify")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Verify failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var verify structs.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verify); err != nil {
		jsonError(w, http.StatusInternalServerError, "Verify JSON error: "+err.Error())
		return
	}

	// Token & Cookie
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

	charID := int64(verify.CharacterID)

	// User upserten
	if err := db2.UpsertUser(charID, verify.CharacterName); err != nil {
		log.Printf("UpsertUser: %v", err)
	}

	// ——— NEU: Zugehörigkeit via Affiliation (frisch) + Details resolven ———
	corpID, corpName, corpTicker, alliID, alliName, alliTicker :=
		esi.FetchCorpAndAlliance(verify.CharacterID)

	// Alliance optional
	var alliPtr *int64
	if alliID != nil && *alliID != 0 {
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

	// Corp + User setzen
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

// /me – leichtgewichtig: nur Verify (keine ESI-Polllawine)
func MeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("char")
	if err != nil || cookie.Value == "" {
		jsonError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	token, ok := esiauth.LoadToken(cookie.Value)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "No token for user")
		return
	}

	client := esiauth.GetOAuthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://esi.evetech.net/verify")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Verify failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var verify structs.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verify); err != nil {
		jsonError(w, http.StatusInternalServerError, "Verify JSON error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, verify)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("char"); err == nil && c.Value != "" {
		esiauth.DeleteToken()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "char",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// /role
func GetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("char")
	if err != nil || c.Value == "" {
		jsonError(w, http.StatusUnauthorized, "Not logged in")
		return
	}
	charID, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Bad char id")
		return
	}

	role, err := db2.GetUserRoles(charID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "DB error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"role": role})
}
