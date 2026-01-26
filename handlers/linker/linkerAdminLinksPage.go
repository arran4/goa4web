package linker

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// AdminLinksPage lists all links grouped by category.
func AdminLinksPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Links"
	LinkerAdminLinksPageTmpl.Handle(w, r, struct{}{})
}

const LinkerAdminLinksPageTmpl tasks.Template = "linker/linkerAdminLinksPage.gohtml"
