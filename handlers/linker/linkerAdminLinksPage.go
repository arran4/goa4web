package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminLinksPage lists all linker items.
func AdminLinksPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Links []*db.GetAllLinkersForIndexRow
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Linker Links"

	queries := cd.Queries()
	rows, err := queries.GetAllLinkersForIndex(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data := Data{Links: rows}
	handlers.TemplateHandler(w, r, "linkerAdminLinksPage.gohtml", data)
}
