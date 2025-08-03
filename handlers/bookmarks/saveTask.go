package bookmarks

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveTask persists changes for a bookmark list.
type SaveTask struct{ tasks.TaskString }

var saveTask = &SaveTask{TaskString: TaskSave}

// ensure SaveTask implements tasks.Task for routing
var _ tasks.Task = (*SaveTask)(nil)

func EditPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	_ = session
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Edit Bookmarks"
	handlers.TemplateHandler(w, r, "editPage.gohtml", struct{}{})
}

func (SaveTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.UpdateBookmarksForLister(r.Context(), db.UpdateBookmarksForListerParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		ListerID:  uid,
		GranteeID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	}); err != nil {
		return fmt.Errorf("update bookmarks fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/bookmarks/mine"}
}
