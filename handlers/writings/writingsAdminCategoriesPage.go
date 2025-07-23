package writings

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

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories          []*db.WritingCategory
		CategoryBreadcrumbs []*db.WritingCategory
		IsAdmin             bool
		IsWriter            bool
		Abstracts           []*db.GetPublicWritingsInCategoryForUserRow
		WritingCategoryID   int32
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.IsAdmin = data.CoreData.HasRole("administrator") && data.CoreData.AdminMode
	data.IsWriter = data.CoreData.HasRole("content writer") || data.IsAdmin

	categoryRows, err := data.CoreData.WritingCategories()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("writingCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	handlers.TemplateHandler(w, r, "categoriesPage.gohtml", data)
}
