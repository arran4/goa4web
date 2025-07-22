package bookmarks

import (
	"database/sql"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// CreateTask creates a new bookmark list for the user.
type CreateTask struct{ tasks.TaskString }

var createTask = &CreateTask{TaskString: TaskCreate}

// ensure CreateTask implements tasks.Task for routing
var _ tasks.Task = (*CreateTask)(nil)

func (CreateTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.CreateBookmarks(r.Context(), db.CreateBookmarksParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/bookmarks/mine", http.StatusTemporaryRedirect)
}
