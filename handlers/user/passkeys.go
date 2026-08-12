package user

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"fmt"
	"github.com/arran4/goa4web/internal/db"
)

func passkeysBeginRegistration(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	wa, err := cd.GetWebAuthn()
	if err != nil {
		http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
		return
	}

	user, err := cd.GetWebAuthnUserByID(cd.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	options, sessionData, err := wa.BeginRegistration(user)
	if err != nil {
		log.Printf("BeginRegistration failed: %v", err)
		http.Error(w, "BeginRegistration failed", http.StatusInternalServerError)
		return
	}

	sess := cd.GetSession()
	sess.Values["webauthn_reg_session"] = sessionData
	if err := sess.Save(r, w); err != nil { log.Printf("Failed to save session: %v", err); http.Error(w, "Failed to save session", http.StatusInternalServerError); return }

	if err := json.NewEncoder(w).Encode(options); err != nil { log.Printf("Failed to encode JSON: %v", err) }
}

func passkeysFinishRegistration(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	wa, err := cd.GetWebAuthn()
	if err != nil {
		http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
		return
	}

	sess := cd.GetSession()
	sessionDataObj, ok := sess.Values["webauthn_reg_session"]
	if !ok {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	sessionData, ok := sessionDataObj.(webauthn.SessionData)
	if !ok {
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	user, err := cd.GetWebAuthnUserByID(cd.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(r.Body)
	if err != nil {
		log.Printf("Parse body failed: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cred, err := wa.CreateCredential(user, sessionData, parsedResponse)
	if err != nil {
		log.Printf("CreateCredential failed: %v", err)
		http.Error(w, "Registration failed", http.StatusBadRequest)
		return
	}

	err = cd.SavePasskey(cred, cd.UserID)
	if err != nil {
		log.Printf("SavePasskey failed: %v", err)
		http.Error(w, "Failed to save passkey", http.StatusInternalServerError)
		return
	}

	delete(sess.Values, "webauthn_reg_session")
	if err := sess.Save(r, w); err != nil { log.Printf("Failed to save session: %v", err) }

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil { log.Printf("Failed to encode JSON: %v", err) }
}

func passkeysDelete(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}
	// For simplicity, just convert to int or assume it's the ID.
	// Wait, let's implement the DB delete passing id as int32
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil { log.Printf("Failed to parse id: %v", err) }
	err := cd.Queries().DeletePasskey(r.Context(), db.DeletePasskeyParams{
		ID:     id,
		UserID: cd.UserID,
	})
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/usr/passkeys", http.StatusSeeOther)
}

func passkeysPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	user, err := cd.GetWebAuthnUserByID(cd.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	cd.PageTitle = "Manage Passkeys"
	if err := cd.ExecuteSiteTemplate(w, r, "domains/user/passkeys.gohtml", user); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
