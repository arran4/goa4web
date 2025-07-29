package writings

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Category   *db.WritingCategory
		Categories []*db.WritingCategory
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cat, err := queries.GetWritingCategory(r.Context(), int32(cid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	cats, err := cd.WritingCategories()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = "Writing Category"
	data := Data{CoreData: cd, Category: cat, Categories: cats}
	handlers.TemplateHandler(w, r, "writingsAdminCategoryPage.gohtml", data)
}
