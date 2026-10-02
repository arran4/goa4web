package privateforum

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// maxResumablePrivateTopicFormBytes bounds persisted URL-encoded submissions.
const maxResumablePrivateTopicFormBytes int64 = 64 << 10

// LimitResumablePrivateTopicPost bounds the only POST body eligible for capture.
func LimitResumablePrivateTopicPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/private/topic/new" {
			r.Body = http.MaxBytesReader(w, r.Body, maxResumablePrivateTopicFormBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// InterceptStalePrivateTopicPost captures one opted-in stale form submission.
func InterceptStalePrivateTopicPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost || r.URL.Path != "/private/topic/new" {
		return false
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		return false
	}
	if err := r.ParseForm(); err != nil {
		return false
	}

	formNonce := r.PostFormValue("resume_nonce")
	if formNonce == "" {
		return false
	}
	formNonceHash := sha256.Sum256([]byte(formNonce))
	formNonceHashHex := hex.EncodeToString(formNonceHash[:])

	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		return false
	}

	if cd.UserID != 0 {
		return false
	}

	browserID := core.GetBrowserID(w, r)
	if browserID == "" {
		return false
	}

	pendingAction, err := cd.PendingAction(r.Context(), formNonceHashHex)
	if err != nil {
		return false
	}

	if pendingAction.BrowserID != browserID || pendingAction.FormData != "" {
		return false
	}

	taskName := r.PostFormValue("task")
	expectedTask := string(TaskPrivateTopicCreate)
	if pendingAction.ActionType != expectedTask || taskName != expectedTask {
		return false
	}

	r.PostForm.Del("gorilla.csrf.Token")
	r.PostForm.Del("resume_nonce")

	storageMap := map[string]any{
		"form": r.PostForm,
		"url":  r.URL.RequestURI(),
	}
	formDataBytes, err := json.Marshal(storageMap)
	if err != nil {
		return false
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return false
	}
	resumeToken := hex.EncodeToString(b)
	resumeTokenHash := sha256.Sum256([]byte(resumeToken))
	resumeTokenHashHex := hex.EncodeToString(resumeTokenHash[:])

	rows, err := cd.CapturePendingAction(r.Context(), db.SystemCapturePendingActionParams{
		FormData: string(formDataBytes),
		ID:       resumeTokenHashHex,
		ID_2:     formNonceHashHex,
	})
	if err != nil || rows == 0 {
		return false
	}

	resumeValues := url.Values{}
	resumeValues.Set("token", resumeToken)
	backURL := "/resume?" + resumeValues.Encode()

	newVals := url.Values{}
	newVals.Set("back", backURL)
	target := "/login?" + newVals.Encode()

	core.DisableCaching(w)
	http.Redirect(w, r, target, http.StatusSeeOther)
	return true
}

// ConsumePrivateTopicFormNonce validates and retires a normal authenticated form attempt.
func ConsumePrivateTopicFormNonce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			rejectPendingAction(w)
			return
		}
		nonce := r.PostFormValue("resume_nonce")
		if nonce == "" {
			rejectPendingAction(w)
			return
		}
		nonceHash := sha256.Sum256([]byte(nonce))
		nonceHashHex := hex.EncodeToString(nonceHash[:])
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil || cd.UserID == 0 {
			rejectPendingAction(w)
			return
		}
		pendingAction, err := cd.PendingAction(r.Context(), nonceHashHex)
		if err != nil || pendingAction.Uid != cd.UserID || pendingAction.BrowserID != core.GetBrowserID(w, r) ||
			pendingAction.ActionType != string(TaskPrivateTopicCreate) || pendingAction.FormData != "" {
			rejectPendingAction(w)
			return
		}
		rows, err := cd.ConsumePendingAction(r.Context(), nonceHashHex)
		if err != nil || rows != 1 {
			rejectPendingAction(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rejectPendingAction(w http.ResponseWriter) {
	core.DisableCaching(w)
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}
