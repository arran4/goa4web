package writings

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// WritingCategoryCreateTask creates a new category.
type WritingCategoryCreateTask struct{ tasks.TaskString }

var writingCategoryCreateTask = &WritingCategoryCreateTask{TaskString: TaskWritingCategoryCreate}

var _ tasks.Task = (*WritingCategoryCreateTask)(nil)

func (WritingCategoryCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		return fmt.Errorf("pcid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.CreateWritingCategory(int32(pcid), name, desc); err != nil {
		return fmt.Errorf("create writing category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
