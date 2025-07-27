package writings

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// CategoryGrantDeleteTask removes a grant from a writing category.
type CategoryGrantDeleteTask struct{ tasks.TaskString }

var categoryGrantDeleteTask = &CategoryGrantDeleteTask{TaskString: TaskCategoryGrantDelete}

var _ tasks.Task = (*CategoryGrantDeleteTask)(nil)

func (CategoryGrantDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	grantID, err := strconv.Atoi(r.PostFormValue("grantid"))
	if err != nil {
		return fmt.Errorf("grant id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := queries.DeleteGrant(r.Context(), int32(grantID)); err != nil {
		log.Printf("DeleteGrant: %v", err)
		return fmt.Errorf("delete grant %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
