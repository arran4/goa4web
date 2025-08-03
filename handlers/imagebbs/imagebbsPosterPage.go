package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func PosterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Posts    []*db.ListImagePostsByPosterForListerRow
		Username string
		IsOffset bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = fmt.Sprintf("Images by %s", username)
	queries := cd.Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("SystemGetUserByUsername Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	rows, err := queries.ListImagePostsByPosterForLister(r.Context(), db.ListImagePostsByPosterForListerParams{
		ListerID:     cd.UserID,
		PosterID:     u.Idusers,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        15,
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("ListImagePostsByPosterForLister Error: %s", err)
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
