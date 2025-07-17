package imagebbs

import (
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *corecorecommon.CoreData }{r.Context().Value(corecorecommon.KeyCoreData).(*corecorecommon.CoreData)}
	hcommon.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
