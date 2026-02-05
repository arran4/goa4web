package linker

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.GetLinkerCategoryLinkCountsRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Link Categories"
	data := Data{}

	categoryRows, err := cd.LinkerCategoryCounts()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("adminCategories Error: %s", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	data.Categories = categoryRows

	LinkerAdminCategoriesPageTmpl.Handle(w, r, data)
}

const LinkerAdminCategoriesPageTmpl tasks.Template = "linker/linkerAdminCategoriesPage.gohtml"
