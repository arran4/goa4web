package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.LinkerCategory
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Categories"
	data := Data{}

	queries := cd.Queries()

	categories, err := queries.GetAllLinkerCategoriesForUser(r.Context(), db.GetAllLinkerCategoriesForUserParams{
		ViewerID:     cd.UserID,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerCategories Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	data.Categories = categories

	LinkerCategoriesPageTmpl.Handle(w, r, data)
}

const LinkerCategoriesPageTmpl handlers.Page = "linker/linkerCategoriesPage.gohtml"
