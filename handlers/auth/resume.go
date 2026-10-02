package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/privateforum"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ResumeInterstitialPageTmpl renders the pending-action confirmation page.
const ResumeInterstitialPageTmpl tasks.Template = "domains/user/resume_interstitial.gohtml"

// ResumePage renders the confirmation interstitial for a bound pending action.
func ResumePage(w http.ResponseWriter, r *http.Request) {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	token := r.URL.Query().Get("token")
	action, _, _, _, err := validatedResumeAction(w, r, cd, token)
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	data := struct {
		Token      string
		ActionType string
	}{Token: token, ActionType: action.ActionType}

	_ = ResumeInterstitialPageTmpl.Handle(w, r, data)
}

// ResumeTaskAction cancels or claims and executes a bound pending action.
func ResumeTaskAction(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	token := r.PostFormValue("token")
	_, targetURL, storedForm, tokenHashHex, err := validatedResumeAction(w, r, cd, token)
	if err != nil {
		return err
	}

	switch r.PostFormValue("operation") {
	case "cancel":
		rows, consumeErr := cd.ConsumePendingAction(r.Context(), tokenHashHex)
		if consumeErr != nil || rows != 1 {
			return handlers.ErrNotFound
		}
		return handlers.RedirectHandler("/private/topic/new")
	case "resume":
	default:
		return handlers.ErrForbidden
	}

	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return handlers.ErrForbidden
	}
	if !cd.HasGrant("privateforum", "topic", "create", 0) {
		return handlers.ErrForbidden
	}

	// Claim before the non-idempotent action. This intentionally provides an
	// at-most-once attempt: validation, database, or process failure after this
	// point can lose the pending submission rather than execute it twice.
	rows, err := cd.ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.RequestURI = targetURL.RequestURI()
	newReq.Method = http.MethodPost
	newReq.Body = http.NoBody
	newReq.Form = storedForm
	newReq.PostForm = storedForm

	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}

func validatedResumeAction(w http.ResponseWriter, r *http.Request, cd *common.CoreData, token string) (*db.PendingAction, *url.URL, url.Values, string, error) {
	if token == "" {
		return nil, nil, nil, "", handlers.ErrNotFound
	}
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])
	action, err := cd.PendingAction(r.Context(), tokenHashHex)
	if err != nil {
		return nil, nil, nil, "", handlers.ErrNotFound
	}
	if action.Uid != cd.UserID || action.BrowserID != core.GetBrowserID(w, r) {
		return nil, nil, nil, "", handlers.ErrForbidden
	}
	if action.ActionType != string(privateforum.TaskPrivateTopicCreate) || action.FormData == "" {
		return nil, nil, nil, "", handlers.ErrForbidden
	}

	var stored struct {
		Form url.Values `json:"form"`
		URL  string     `json:"url"`
	}
	if err := json.Unmarshal([]byte(action.FormData), &stored); err != nil {
		return nil, nil, nil, "", fmt.Errorf("decode pending action: %w", handlers.ErrForbidden)
	}
	targetURL, err := url.Parse(stored.URL)
	if err != nil || targetURL.IsAbs() || targetURL.Host != "" || targetURL.Opaque != "" ||
		targetURL.User != nil || targetURL.Path != "/private/topic/new" || targetURL.Fragment != "" {
		return nil, nil, nil, "", handlers.ErrForbidden
	}
	return action, targetURL, stored.Form, tokenHashHex, nil
}

// ResumeTask adapts pending-action confirmation to the task handler.
type ResumeTask struct {
	tasks.TaskString
}

func (ResumeTask) Action(w http.ResponseWriter, r *http.Request) any {
	return ResumeTaskAction(w, r)
}
