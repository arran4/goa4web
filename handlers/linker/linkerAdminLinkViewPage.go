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
	"github.com/gorilla/mux"
)

// AdminLinkPage displays a linker item overview.
func AdminLinkPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	link, id, err := cd.SelectedAdminLinkerItem(r)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	cd.PageTitle = fmt.Sprintf("Link %d", id)
	data := struct {
		Link *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	}{Link: link}
	handlers.TemplateHandler(w, r, "linkerAdminLinkPage.gohtml", data)
}

// AdminLinkGrantsPage displays grants for a linker item.
func AdminLinkGrantsPage(w http.ResponseWriter, r *http.Request) {
	handlers.TemplateHandler(w, r, "linkerAdminLinkGrantsPage.gohtml", struct{}{})
}

// AdminLinkCommentsPage redirects to the public comments view for the link.
func AdminLinkCommentsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.Redirect(w, r, "/linker/comments/"+vars["link"], http.StatusTemporaryRedirect)
}
