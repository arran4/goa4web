package user

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	// handlers.TemplateHandler(w, r, "loginAttemptsPage.gohtml", struct{}{})
	_ = AdminLoginAttemptsPage.Handle(w, r, struct{}{})
}

const AdminLoginAttemptsPage tasks.Template = "domains/admin/loginAttemptsPage.gohtml"
