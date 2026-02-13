package bookmarks

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// CreateTask creates a new bookmark list.
type CreateTask struct{ tasks.TaskString }

var createTask = &CreateTask{TaskString: TaskCreate}

// ensure CreateTask implements tasks.Task for routing
var _ tasks.Task = (*CreateTask)(nil)

func (CreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("text")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
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
