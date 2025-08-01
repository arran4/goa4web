package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminRevisionHistoryPage shows all revisions for a FAQ entry.
func AdminRevisionHistoryPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	faq, err := queries.GetFAQByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	revs, _ := queries.GetFAQRevisionsForAdmin(r.Context(), int32(id))
	type Data struct {
		*common.CoreData
		Faq       *db.Faq
		Revisions []*db.FaqRevision
	}
	data := Data{
		CoreData:  r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Faq:       faq,
		Revisions: revs,
	}
	cd := data.CoreData
	cd.PageTitle = fmt.Sprintf("FAQ %d History", id)
	handlers.TemplateHandler(w, r, "adminFaqRevisionPage.gohtml", data)
}
