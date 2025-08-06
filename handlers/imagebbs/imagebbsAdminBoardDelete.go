package imagebbs

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

// DeleteBoardTask removes an image board.
type DeleteBoardTask struct{ tasks.TaskString }

var deleteBoardTask = &DeleteBoardTask{TaskString: TaskDeleteBoard}

var _ tasks.Task = (*DeleteBoardTask)(nil)

func (DeleteBoardTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["board"])
	if bid == 0 {
		return handlers.ErrRedirectOnSamePageHandler(handlers.ErrBadRequest)
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.Queries().AdminDeleteImageBoard(r.Context(), int32(bid)); err != nil {
		return fmt.Errorf("delete image board %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/imagebbs/boards"}
}
