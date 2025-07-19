package faq

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AnswerTask struct{ tasks.TaskString }
type RemoveQuestionTask struct{ tasks.TaskString }

var answerTask = &AnswerTask{TaskString: TaskAnswer}
var removeQuestionTask = &RemoveQuestionTask{TaskString: TaskRemoveRemove}

// Implementing these interfaces means answering a FAQ automatically notifies
// the original asker and the administrators. From a user's perspective this
// ensures they are kept in the loop once their question is addressed.
var _ notif.AdminEmailTemplateProvider = (*AnswerTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*AnswerTask)(nil)

func (AnswerTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("faqAnsweredEmail")
}

func (AnswerTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("faq_answered")
	return &v
}

func (AnswerTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("faqAnsweredEmail")
}

func (AnswerTask) SelfInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("faq_answered")
	return &v
}

func (AnswerTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskAnswer)(r, m)
}

func (RemoveQuestionTask) Match(req *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRemoveRemove)(req, m)
}

func AdminAnswerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories []*db.FaqCategory
		Rows       []*db.Faq
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	catrows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Categories = catrows

	rows, err := queries.GetFAQUnansweredQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "adminAnswerPage.gohtml", data)
}

func (AnswerTask) Action(w http.ResponseWriter, r *http.Request) {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (RemoveQuestionTask) Action(w http.ResponseWriter, r *http.Request) {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.DeleteFAQ(r.Context(), int32(faq)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
