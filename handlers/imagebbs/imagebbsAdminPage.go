package imagebbs

import (
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *handlers.CoreData }{r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData)}
	handlers.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
