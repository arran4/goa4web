package auth

import (
	"encoding/json"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"log"
	"net/http"
	"time"
)

func loginPasskeyBegin(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	username := r.URL.Query().Get("username")

	if cd.WebAuthn == nil {
		http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
		return
	}

	user, err := cd.GetWebAuthnUser(username)
	if err != nil {
		// Log the error but don't leak info
		log.Printf("GetWebAuthnUser failed: %v", err)
		http.Error(w, "User not found or no passkeys", http.StatusNotFound)
		return
	}

	options, sessionData, err := cd.WebAuthn.BeginLogin(user)
	if err != nil {
		log.Printf("BeginLogin failed: %v", err)
		http.Error(w, "BeginLogin failed", http.StatusInternalServerError)
		return
	}

	sess := cd.GetSession()
	sess.Values["webauthn_session"] = sessionData
	sess.Values["webauthn_uid"] = user.Idusers
	if err := sess.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}

func loginPasskeyFinish(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if cd.WebAuthn == nil {
		http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
		return
	}

	sess := cd.GetSession()
	sessionDataObj, ok := sess.Values["webauthn_session"]
	if !ok {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	sessionData, ok := sessionDataObj.(webauthn.SessionData)
	if !ok {
		// Try to unmarshal if it's a map (sessions library might encode structs as map[string]interface{})
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	uid, ok := sess.Values["webauthn_uid"].(int32)
	if !ok {
		http.Error(w, "No active user", http.StatusBadRequest)
		return
	}

	user, err := cd.GetWebAuthnUserByID(uid)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		log.Printf("Parse body failed: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cred, err := cd.WebAuthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
		return user, nil
	}, sessionData, parsedResponse)

	if err != nil {
		log.Printf("ValidateLogin passkey failed for user %d: %v", uid, err)
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

	// Update sign count
	err = cd.Queries().UpdatePasskeySignCount(r.Context(), common.ToUpdatePasskeySignCountParams(cred))
	if err != nil {
		log.Printf("Failed to update sign count: %v", err)
	}

	delete(sess.Values, "webauthn_session")
	delete(sess.Values, "webauthn_uid")

	sess.Values["UID"] = int32(user.Idusers)
	sess.Values["LoginTime"] = time.Now().Unix()
	sess.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("login success passkey uid=%d session=%s", user.Idusers, handlers.HashSessionID(sess.ID))
	}

	if err := sess.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}
