package imagebbs

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *common.CoreData }{r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	handlers.SetPageTitle(r, "Image Board Admin")
	handlers.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
