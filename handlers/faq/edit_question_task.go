package faq

import (
	"fmt"
	"net/http"
	"strconv"


	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// EditQuestionTask modifies an existing FAQ entry.
type EditQuestionTask struct{ tasks.TaskString }

var editQuestionTask = &EditQuestionTask{TaskString: TaskEdit}
var _ tasks.Task = (*EditQuestionTask)(nil)

func (EditQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskEdit)(r, m)
}

func (EditQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	description := r.PostFormValue("description")
	version := r.PostFormValue("version")
	priority, _ := strconv.Atoi(r.PostFormValue("priority"))
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)

	if err := cd.UpdateFAQQuestion(question, answer, description, version, int32(category), int32(faq), uid, int32(priority)); err != nil {
		return fmt.Errorf("update faq question fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
