package imagebbs

import (
	corecommon "github.com/arran4/goa4web/core/common"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *corecommon.CoreData }{r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData)}
	handlers.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
