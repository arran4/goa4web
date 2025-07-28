package linker

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

func UserPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Links     []*db.GetLinkerItemsByUserDescendingForUserRow
		Username  string
		HasOffset bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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

	uid := r.Context().Value(consts.KeyCoreData).(*common.CoreData).UserID
	rows, err := queries.GetLinkerItemsByUserDescendingForUser(r.Context(), db.GetLinkerItemsByUserDescendingForUserParams{
		ViewerID:     uid,
		UserID:       u.Idusers,
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		Limit:        15,
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetLinkerItemsByUserDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{CoreData: cd, Username: username, HasOffset: offset != 0}
	cd.PageTitle = fmt.Sprintf("Links by %s", username)
	for _, row := range rows {
		if !cd.HasGrant("linker", "link", "see", row.Idlinker) {
			continue
		}
		data.Links = append(data.Links, row)
	}

	handlers.TemplateHandler(w, r, "linkerPage", data)
}
