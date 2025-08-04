package user

import (
	"net/http"

	"github.com/arran4/goa4web/handlers"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	handlers.TemplateHandler(w, r, "loginAttemptsPage.gohtml", struct{}{})
}
