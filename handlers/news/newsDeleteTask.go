package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteNewsPostTask removes a news post.
type DeleteNewsPostTask struct{ tasks.TaskString }

var deleteNewsPostTask = &DeleteNewsPostTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteNewsPostTask)(nil)

func (DeleteNewsPostTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		return fmt.Errorf("post id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.DeleteNewsPost(int32(pid)); err != nil {
		return fmt.Errorf("delete news post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd != nil {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["NewsPostID"] = pid
		}
	}
	return nil
}
