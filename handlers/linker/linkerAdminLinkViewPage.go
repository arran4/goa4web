package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminLinkViewPage displays information about a linker item.
func adminLinkViewPage(w http.ResponseWriter, r *http.Request) {
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

	if link.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Link: %s", link.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Link %d", id)
	}

	data := struct {
		Link *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	}{
		Link: link,
	}

	LinkerAdminLinkViewPageTmpl.Handle(w, r, data)
}

const LinkerAdminLinkViewPageTmpl tasks.Template = "linker/adminLinkViewPage.gohtml"
