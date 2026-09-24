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
	"github.com/arran4/goa4web/internal/tasks"
)

var ResumeInterstitialPageTmpl tasks.Template = "domains/user/resume_interstitial.gohtml"

func ResumePage(w http.ResponseWriter, r *http.Request) {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	browserID := core.GetBrowserID(w, r)

	action, err := cd.Queries().GetPendingAction(r.Context(), tokenHashHex)
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	if action.Uid != cd.UserID || action.BrowserID != browserID {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data := struct {
		Token      string
		ActionType string
	}{Token: token, ActionType: action.ActionType}

	_ = ResumeInterstitialPageTmpl.Handle(w, r, data)
}

func ResumeTaskAction(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	token := r.PostFormValue("token")
	if token == "" {
		return handlers.ErrNotFound
	}
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	browserID := core.GetBrowserID(w, r)

	action, err := cd.Queries().GetPendingAction(r.Context(), tokenHashHex)
	if err != nil {
		return handlers.ErrNotFound
	}

	if action.Uid != cd.UserID || action.BrowserID != browserID {
		return handlers.ErrForbidden
	}

	if action.ActionType != string(privateforum.TaskPrivateTopicCreate) {
		return handlers.ErrForbidden
	}

	var storageMap struct {
		Form url.Values `json:"form"`
		URL  string     `json:"url"`
	}
	if err := json.Unmarshal([]byte(action.FormData), &storageMap); err != nil {
		return fmt.Errorf("invalid form data")
	}

	targetURL, err := url.Parse(storageMap.URL)
	if err != nil || targetURL.IsAbs() || targetURL.Host != "" || targetURL.Path != "/private/topic/new" {
		return handlers.ErrForbidden
	}

	// 1. Explicitly check current authorization BEFORE consuming the token.
	// We want to preserve the token if the user is legitimate but simply unauthorized right now.
	if !cd.HasGrant("privateforum", "topic", "see", 0) || !cd.HasGrant("privateforum", "topic", "create", 0) {
		return handlers.ErrForbidden
	}

	// 2. Consume atomically AFTER authorization checks to ensure exactly-once semantics.
	// Since we don't have global explicit multi-statement transactions in the app's framework
	// for arbitrary actions, we at least prevent duplicate submission concurrency natively via
	// the Consume SQL query, and error cleanly on partial failure without duplicate topics.
	rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.Method = http.MethodPost
	newReq.PostForm = storageMap.Form

	// Execute action directly, returning its result properly to the outer TaskHandler
	// since we already explicitly checked the authorization constraint above.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}

type ResumeTask struct {
	tasks.TaskString
}

var resumeTask = ResumeTask{TaskString: "resumeTask"}

func (ResumeTask) Action(w http.ResponseWriter, r *http.Request) any {
	return ResumeTaskAction(w, r)
}
