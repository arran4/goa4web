package auth

import (
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

type loginFormHandler struct{ msg string }

func (l loginFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, l.msg, "")
}

var _ http.Handler = (*loginFormHandler)(nil)

type redirectBackPageHandler struct {
	BackURL string
	Method  string
	Values  url.Values
}

func (h redirectBackPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if h.Method == "" || h.Method == http.MethodGet {
		targetURL := h.BackURL
		if len(h.Values) > 0 {
			if u, err := url.Parse(targetURL); err == nil {
				q := u.Query()
				for k, vs := range h.Values {
					for _, v := range vs {
						q.Add(k, v)
					}
				}
				u.RawQuery = q.Encode()
				targetURL = u.String()
			}
		}
		rdh := handlers.RefreshDirectHandler{TargetURL: targetURL}
		cd.AutoRefresh = rdh.Content()
		TaskDoneAutoRefreshPageTmpl.Handle(w, r, rdh)
		return
	}

	type Data struct {
		BackURL string
		Method  string
		Values  url.Values
	}
	if err := RedirectBackPageTmpl.Handle(w, r, Data(h)); err != nil {
		log.Printf("Template Error: %s", err)
		handlers.RenderErrorPage(w, r, err)
	}
}

const (
	TaskDoneAutoRefreshPageTmpl tasks.Template = "taskDoneAutoRefreshPage.gohtml"
	RedirectBackPageTmpl        tasks.Template = "redirectBackPage.gohtml"
)

var _ http.Handler = (*redirectBackPageHandler)(nil)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg, noticeMsg string) {
	handlers.SetNoCacheHeaders(w)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.SetCurrentError(errMsg)
	cd.SetCurrentNotice(noticeMsg)
	type Data struct {
		Code    string
		Back    string
		BackSig string
		BackTS  string
		Method  string
		Data    string
	}
	handlers.SetPageTitle(r, "Login")
	data := Data{
		Code:    r.FormValue("code"),
		Back:    cd.SanitizeBackURL(r, r.FormValue("back")),
		BackSig: r.FormValue("back_sig"),
		BackTS:  r.FormValue("back_ts"),
		Method:  r.FormValue("method"),
		Data:    r.FormValue("data"),
	}
	LoginPageTmpl.Handle(w, r, data)
}

const LoginPageTmpl tasks.Template = "loginPage.gohtml"
