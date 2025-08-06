package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// adminLinkGrantDeleteTask removes a grant from a linker item.
type linkGrantDeleteTask struct{ tasks.TaskString }

var AdminLinkGrantDeleteTask = &linkGrantDeleteTask{TaskString: TaskCategoryGrantDelete}

var _ tasks.Task = (*linkGrantDeleteTask)(nil)

func (linkGrantDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	grantID, _ := strconv.Atoi(r.PostFormValue("grantid"))
	if err := queries.AdminDeleteGrant(r.Context(), int32(grantID)); err != nil {
		return fmt.Errorf("delete grant fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
