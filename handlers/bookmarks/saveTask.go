package bookmarks

import (
	"database/sql"
	"errors"
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

func (SaveTask) Page(w http.ResponseWriter, r *http.Request) {
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

	handlers.TemplateHandler(w, r, "editPage.gohtml", data)
}

func (SaveTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return nil
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.UpdateBookmarks(r.Context(), db.UpdateBookmarksParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}

	http.Redirect(w, r, "/bookmarks/mine", http.StatusTemporaryRedirect)

	return nil
}
