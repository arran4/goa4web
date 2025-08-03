package bookmarks

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
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
	type Data struct {
		*common.CoreData
		BookmarkContent string
		Bid             interface{}
	}

	data := Data{
		CoreData:        r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		BookmarkContent: "Category: Example 1\nhttp://www.google.com.au Google\nColumn\nCategory: Example 2\nhttp://www.google.com.au Google\nhttp://www.google.com.au Google\n",
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	_ = session
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	bookmarks, err := cd.Bookmarks()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("error getBookmarksForUser: %s", err)
			http.Error(w, "ERROR", 500)
			return
		}
	} else {
		data.BookmarkContent = bookmarks.List.String
		data.Bid = bookmarks.Idbookmarks
	}

	cd.PageTitle = "Edit Bookmarks"
	handlers.TemplateHandler(w, r, "editPage.gohtml", data)
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
