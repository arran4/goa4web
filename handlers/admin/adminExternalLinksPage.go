package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminExternalLinksPage lists cached external links.
func AdminExternalLinksPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Links []*db.ExternalLink
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "External Links"
	queries := cd.Queries()
	rows, err := queries.AdminListExternalLinks(r.Context(), db.AdminListExternalLinksParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list external links: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := Data{CoreData: cd, Links: rows}
	handlers.TemplateHandler(w, r, "admin/externalLinksPage.gohtml", data)
}
