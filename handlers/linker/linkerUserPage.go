package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func UserPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Links     []*db.GetLinkerItemsByUserDescendingForUserRow
		Username  string
		HasOffset bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("SystemGetUserByUsername Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{Username: username, HasOffset: offset != 0}
	cd.PageTitle = fmt.Sprintf("Links by %s", username)
	for _, row := range rows {
		if !cd.HasGrant("linker", "link", "see", row.ID) {
			continue
		}
		data.Links = append(data.Links, row)
	}

	LinkerUserPageTmpl.Handle(w, r, data)
}

const LinkerUserPageTmpl handlers.Page = "linker/linkerPage.gohtml"
