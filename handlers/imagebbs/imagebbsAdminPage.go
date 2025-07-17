package imagebbs

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *common.CoreData }{r.Context().Value(common.KeyCoreData).(*common.CoreData)}
	handlers.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
