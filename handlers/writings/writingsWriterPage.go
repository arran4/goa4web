package writings

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

func WriterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Abstracts []*db.ListPublicWritingsByUserForListerRow
		Username  string
		IsOffset  bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = fmt.Sprintf("Writer: %s", username)

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

	rows, err := cd.WriterWritings(u.Idusers, r)
	if err != nil {
		log.Printf("WriterWritings: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	data := Data{
		Username: username,
		IsOffset: offset != 0,
	}
	for _, row := range rows {
		if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, row)
	}

	handlers.TemplateHandler(w, r, "writerPage.gohtml", data)
}
