package bookmarks

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type SaveTask struct{ tasks.BasicTaskEvent }

var saveTask = SaveTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskSave,
		Match:     tasks.HasTask(TaskSave),
	},
}

type CreateTask struct{ tasks.BasicTaskEvent }

var createTask = CreateTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskCreate,
		Match:     tasks.HasTask(TaskCreate),
	},
}

func (saveTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		BookmarkContent string
		Bid             interface{}
	}

	data := Data{
		CoreData:        r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
		BookmarkContent: "Category: Example 1\nhttp://www.google.com.au Google\nColumn\nCategory: Example 2\nhttp://www.google.com.au Google\nhttp://www.google.com.au Google\n",
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	_ = session
	cd := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData)
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

func (saveTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
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
		return
	}

	http.Redirect(w, r, "/bookmarks/mine", http.StatusTemporaryRedirect)

}

func (createTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
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
