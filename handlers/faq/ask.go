package faq

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AskTask struct{ tasks.TaskString }

var askTask = &AskTask{TaskString: TaskAsk}

var _ tasks.Task = (*AskTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AskTask)(nil)

func (AskTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationFaqAskEmail")
}

func (AskTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationFaqAskEmail")
	return &v
}

func (AskTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskAsk)(r, m)
}

func (AskTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Ask a Question"
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: cd.PreferredLanguageID(cd.Config.DefaultLanguage),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "askPage.gohtml", data)
}

func (AskTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("faq", "question", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}

	if err := queries.CreateFAQQuestion(r.Context(), db.CreateFAQQuestionParams{
		Question: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
	}); err != nil {
		return fmt.Errorf("faq fetch fail: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	evt := cd.Event()
	evt.Path = "/admin/faq"
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	evt.Data["Question"] = text

	// The BusWorker sends notifications based on event metadata.
	// Setting Admin=true signals administrators should be alerted.

	return handlers.RefreshDirectHandler{TargetURL: "/faq"}
}
