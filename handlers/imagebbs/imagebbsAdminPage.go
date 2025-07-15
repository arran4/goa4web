package imagebbs

import (
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	hcommon.TemplateHandler("imagebbsAdminPage").ServeHTTP(w, r)
}
