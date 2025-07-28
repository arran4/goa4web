package writings

import (
	"database/sql"
	"errors"
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
		*common.CoreData
		Abstracts []*db.GetPublicWritingsByUserForViewerRow
		Username  string
		IsOffset  bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]
	handlers.SetPageTitlef(r, "Writer: %s", username)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
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

	rows, err := cd.WriterWritings(u.Idusers, r)
	if err != nil {
		log.Printf("WriterWritings: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: cd,
		Username: username,
		IsOffset: offset != 0,
	}
	for _, row := range rows {
		if !data.CoreData.HasGrant("writing", "article", "see", row.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, row)
	}

	handlers.TemplateHandler(w, r, "writerPage.gohtml", data)
}
