package faq

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AnswerTask submits an answer in the FAQ admin interface.
type AnswerTask struct{ tasks.TaskString }

// RemoveQuestionTask deletes a question from the FAQ list.
type RemoveQuestionTask struct{ tasks.TaskString }

var answerTask = &AnswerTask{TaskString: TaskAnswer}

var _ tasks.Task = (*AnswerTask)(nil)
var removeQuestionTask = &RemoveQuestionTask{TaskString: TaskRemoveRemove}

var _ tasks.Task = (*RemoveQuestionTask)(nil)

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

func (AnswerTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
	return nil
}

func (RemoveQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.DeleteFAQ(r.Context(), int32(faq)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
	return nil
}
