package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"

	"github.com/gorilla/mux"
)

// BloggerPostsPage shows the posts written by a specific blogger.
func BloggerPostsPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntriesForUserDescendingLanguagesRow
		EditUrl string
	}
	type Data struct {
		*common.CoreData
		Rows     []*BlogRow
		IsOffset bool
		UID      string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, ok := cd.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := cd.Queries()

	bu, err := queries.GetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("GetUserByUsername Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buid := bu.Idusers

	rows, err := queries.GetBlogEntriesForUserDescendingLanguages(r.Context(), db.GetBlogEntriesForUserDescendingLanguagesParams{
		UsersIdusers:  buid,
		ViewerIdusers: uid,
		Limit:         15,
		Offset:        int32(offset),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Query Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		IsOffset: offset != 0,
		UID:      strconv.Itoa(int(buid)),
	}

	for _, row := range rows {
		editUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/blogs/blog/%d/edit", row.Idblogs)
		}
		data.Rows = append(data.Rows, &BlogRow{
			GetBlogEntriesForUserDescendingLanguagesRow: row,
			EditUrl: editUrl,
		})
	}

	handlers.TemplateHandler(w, r, "bloggerPostsPage.gohtml", data)
}
