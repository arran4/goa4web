package routes

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// CreateTask creates a new bookmark list.
type createTask struct{ tasks.TaskString }

var CreateTask = &createTask{TaskString: TaskCreate}

// ensure CreateTask implements tasks.Task for routing
var _ tasks.Task = (*createTask)(nil)

func (createTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("text")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if err := cd.CreateBookmark(db.CreateBookmarksForListerParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	}); err != nil {
		return fmt.Errorf("create bookmarks fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/bookmarks/mine"}
}
