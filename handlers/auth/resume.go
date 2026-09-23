package auth

import (
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

	browserID := core.GetBrowserID(w, r)

	action, err := cd.Queries().GetPendingAction(r.Context(), token)
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

	browserID := core.GetBrowserID(w, r)

	action, err := cd.Queries().GetPendingAction(r.Context(), token)
	if err != nil {
		return handlers.ErrNotFound
	}

	if action.Uid != cd.UserID || action.BrowserID != browserID {
		return handlers.ErrForbidden
	}

	rows, err := cd.Queries().ConsumePendingAction(r.Context(), token)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	var formData map[string][]string
	if err := json.Unmarshal([]byte(action.FormData), &formData); err != nil {
		return fmt.Errorf("invalid form data")
	}

	u, _ := url.Parse("/private/topic/new")
	newReq := r.Clone(r.Context())
	newReq.URL = u
	newReq.Method = http.MethodPost
	newReq.PostForm = formData

	handlers.TaskHandler(privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate})(w, newReq)
	return nil
}

type ResumeTask struct {
	tasks.TaskString
}

var resumeTask = ResumeTask{TaskString: "resumeTask"}

func (ResumeTask) Action(w http.ResponseWriter, r *http.Request) any {
	ResumeTaskAction(w, r)
	return nil
}
