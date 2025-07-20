package imagebbs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func PosterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Posts    []*db.GetImagePostsByUserDescendingForUserRow
		Username string
		IsOffset bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	rows, err := queries.GetImagePostsByUserDescendingForUser(r.Context(), db.GetImagePostsByUserDescendingForUserParams{
		ViewerID:     cd.UserID,
		UserID:       u.Idusers,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        15,
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetImagePostsByUserDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filtered := rows

	data := Data{
		CoreData: cd,
		Posts:    filtered,
		Username: username,
		IsOffset: offset != 0,
	}

	handlers.TemplateHandler(w, r, "posterPage.gohtml", data)
}
