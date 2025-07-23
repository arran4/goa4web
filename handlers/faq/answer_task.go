package faq

import (
	"database/sql"
	"fmt"
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

var answerTask = &AnswerTask{TaskString: TaskAnswer}

var _ tasks.Task = (*AnswerTask)(nil)
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

func (AnswerTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		return fmt.Errorf("faq update fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
