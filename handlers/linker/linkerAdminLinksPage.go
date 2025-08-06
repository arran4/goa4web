package linker

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// AdminLinksPage lists all links grouped by category.
func AdminLinksPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Links"
	handlers.TemplateHandler(w, r, "linkerAdminLinksPage.gohtml", struct{}{})
}
