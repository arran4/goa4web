package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecorecommon.CoreData
		Categories []*db.LinkerCategory
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	categories, err := queries.GetAllLinkerCategories(r.Context())
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

	common.TemplateHandler(w, r, "categoriesPage.gohtml", data)
}
