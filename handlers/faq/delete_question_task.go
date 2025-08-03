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

// DeleteQuestionTask removes a question from the FAQ.
type DeleteQuestionTask struct{ tasks.TaskString }

var deleteQuestionTask = &DeleteQuestionTask{TaskString: TaskRemoveRemove}
var _ tasks.Task = (*DeleteQuestionTask)(nil)

func (DeleteQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRemoveRemove)(r, m)
}

func (DeleteQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	if err := queries.AdminDeleteFAQ(r.Context(), int32(faq)); err != nil {
		return fmt.Errorf("delete faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
