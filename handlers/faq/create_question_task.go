package faq

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// CreateQuestionTask creates a new FAQ entry.
type CreateQuestionTask struct{ tasks.TaskString }

var createQuestionTask = &CreateQuestionTask{TaskString: TaskCreate}
var _ tasks.Task = (*CreateQuestionTask)(nil)

func (CreateQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreate)(r, m)
}

func (CreateQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	priority, _ := strconv.Atoi(r.PostFormValue("priority"))
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("faq", "question", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if _, err := cd.CreateFAQQuestion(common.CreateFAQQuestionParams{
		Question:   question,
		Answer:     answer,
		CategoryID: int32(category),
		WriterID:   uid,
		LanguageID: 1,
		Priority:   int32(priority),
	}); err != nil {
		return fmt.Errorf("insert faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
