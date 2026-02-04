package faq

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"

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
var _ tasks.EmailTemplatesRequired = (*AskTask)(nil)

func (AskTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationFaqAsk.EmailTemplates(), true
}

func (AskTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationFaqAsk.NotificationTemplate()
	return &v
}

func (AskTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(AskPageTmpl)},
		EmailTemplateAdminNotificationFaqAsk.RequiredTemplates()...)
}

func (AskTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskAsk)(r, m)
}

func (AskTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Ask a Question"
	data := Data{
		SelectedLanguageId: cd.PreferredLanguageID(cd.Config.DefaultLanguage),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Languages = languageRows

	if _, err := cd.FAQCategories(); err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	AskPageTmpl.Handle(w, r, data)
}

const AskPageTmpl tasks.Template = "faq/askPage.gohtml"

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

	if err := queries.CreateFAQQuestionForWriter(r.Context(), db.CreateFAQQuestionForWriterParams{
		Question:   sql.NullString{String: text, Valid: true},
		WriterID:   uid,
		LanguageID: sql.NullInt32{Int32: int32(languageId), Valid: languageId != 0},
		GranteeID:  sql.NullInt32{Int32: uid, Valid: true},
	}); err != nil {
		return fmt.Errorf("faq fetch fail: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	evt := cd.Event()
	path := "/admin/faq/questions"
	evt.Path = path
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.BaseURL != "" {
		page = strings.TrimRight(cfg.BaseURL, "/") + path
	}
	evt.Data["URL"] = page
	evt.Data["Question"] = text

	// The BusWorker sends notifications based on event metadata.
	// Setting Admin=true signals administrators should be alerted.

	return handlers.RefreshDirectHandler{TargetURL: "/faq"}
}
