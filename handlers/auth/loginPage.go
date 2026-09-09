package auth

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

type loginFormHandler struct{ msg string }

func (l loginFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, l.msg, "")
}

var _ http.Handler = (*loginFormHandler)(nil)

const (
	TaskDoneAutoRefreshPageTmpl tasks.Template = "pages/misc/taskDoneAutoRefreshPage.gohtml"
)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg, noticeMsg string) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.SetCurrentError(errMsg)
	cd.SetCurrentNotice(noticeMsg)
	type Data struct {
		Code    string
		Back    string
		Method  string
	}
	handlers.SetPageTitle(r, "Login")
	backURL, _ := cd.SanitizeBackURL(r, r.FormValue("back"))
	data := Data{
		Code:    r.FormValue("code"),
		Back:    backURL,
		Method:  r.FormValue("method"),
	}
	_ = LoginPageTmpl.Handle(w, r, data)
}

const LoginPageTmpl tasks.Template = "pages/auth/loginPage.gohtml"
