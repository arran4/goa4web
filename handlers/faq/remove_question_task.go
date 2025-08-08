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

// RemoveQuestionTask removes a question from a FAQ category.
type RemoveQuestionTask struct{ tasks.TaskString }

var removeQuestionTask = &RemoveQuestionTask{TaskString: TaskRemoveRemove}

var _ tasks.Task = (*RemoveQuestionTask)(nil)

func (RemoveQuestionTask) Match(req *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRemoveRemove)(req, m)
}

func (RemoveQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	questionID, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	categoryID, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := cd.SelectedQuestionFromCategory(int32(questionID), int32(categoryID)); err != nil {
		return fmt.Errorf("delete faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
