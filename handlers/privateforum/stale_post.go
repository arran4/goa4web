package privateforum

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/middleware/csrf"
)

func init() {
	csrf.StalePostInterceptor = interceptStalePost
}

func interceptStalePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024*64) // 64KB limit
	if err := r.ParseForm(); err != nil {
		return false
	}

	formNonce := r.PostFormValue("resume_nonce")
	if formNonce == "" {
		log.Printf("stale post intercepted failed")
		return false
	}

	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		log.Printf("stale post intercepted failed")
		return false
	}

	browserID := core.GetBrowserID(w, r)
	if browserID == "" {
		log.Printf("stale post intercepted failed")
		return false
	}

	pendingAction, err := cd.Queries().GetPendingAction(r.Context(), formNonce)
	if err != nil {
		log.Printf("GetPendingAction error: %v", err)
		return false
	}

	if pendingAction.BrowserID != browserID {
		log.Printf("browser ID mismatch. expected %s, got %s", pendingAction.BrowserID, browserID)
		return false
	}

	r.PostForm.Del("gorilla.csrf.Token")
	r.PostForm.Del("resume_nonce")

	formDataBytes, err := json.Marshal(r.PostForm)
	if err != nil {
		log.Printf("stale post intercepted failed")
		return false
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("stale post intercepted failed")
		return false
	}
	resumeToken := hex.EncodeToString(b)

	if _, err := cd.Queries().UpdatePendingActionData(r.Context(), db.UpdatePendingActionDataParams{
		FormData: string(formDataBytes),
		ID:       resumeToken,
		ID_2:     formNonce,
	}); err != nil {
		log.Printf("stale post intercepted failed")
		return false
	}

	backURL := "/resume?token=" + resumeToken
	session, _ := core.Store.Get(r, core.SessionName)

	newVals := url.Values{}
	newVals.Set("back", backURL)
	target := "/login?" + newVals.Encode()

	core.DisableCaching(w)
	if session != nil {
		_ = session.Save(r, w)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
	return true
}
