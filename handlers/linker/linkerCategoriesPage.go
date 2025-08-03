package linker

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories []*db.LinkerCategory
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "Categories"

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	cd := data.CoreData
	categories, err := queries.ListLinkerCategoriesForLister(r.Context(), db.ListLinkerCategoriesForListerParams{
		ListerID:     cd.UserID,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        int32(cd.PageSize()),
		Offset:       0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categories

	handlers.TemplateHandler(w, r, "linkerCategoriesPage.gohtml", data)
}
