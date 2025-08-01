package admin

import (
	"database/sql"
	"errors"
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
	rows, err := queries.ListExternalLinksForAdmin(r.Context(), db.ListExternalLinksForAdminParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list external links: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := Data{CoreData: cd, Links: rows}
	handlers.TemplateHandler(w, r, "admin/externalLinksPage.gohtml", data)
}
