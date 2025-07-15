package imagebbs

import (
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *hcommon.CoreData }{r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)}
	hcommon.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
