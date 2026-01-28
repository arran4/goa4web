package writings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

var _ tasks.Task = (*SubmitWritingTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*SubmitWritingTask)(nil)
var _ notif.GrantsRequiredProvider = (*SubmitWritingTask)(nil)
var _ tasks.EmailTemplatesRequired = (*SubmitWritingTask)(nil)

func (SubmitWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleAddPage(w, r) }

func (SubmitWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["category"])
	if err != nil {
		return fmt.Errorf("category id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	private, err := strconv.ParseBool(r.PostFormValue("isitprivate"))
	if err != nil {
		return fmt.Errorf("private parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")
	uid, _ := session.Values["UID"].(int32)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	articleID, err := cd.CreateWriting(int32(categoryID), int32(languageID), title, abstract, body, private)
	if err != nil {
		return fmt.Errorf("create writing fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if articleID == 0 {
		return fmt.Errorf("create writing deny %w", handlers.ErrRedirectOnSamePageHandler(handlers.ErrForbidden))
	}

	var author string
	queries := cd.Queries()
	if u, err := queries.SystemGetUserByID(r.Context(), uid); err == nil {
		author = u.Username.String
	} else {
		return fmt.Errorf("get user fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Title"] = title
		evt.Data["Author"] = author
		evt.Data["target"] = notif.Target{Type: "writing", ID: int32(articleID)}
	}

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: int32(articleID), Text: fullText}
	}

	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", articleID))
}

func (SubmitWritingTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateWriting.EmailTemplates(), true
}

func (SubmitWritingTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateWriting.NotificationTemplate()
	return &s
}

func (SubmitWritingTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(WritingsArticleAddPageTmpl)},
		append(EmailTemplateWriting.RequiredTemplates(), NotificationTemplateWriting.RequiredTemplates()...)...)
}

func (SubmitWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}
