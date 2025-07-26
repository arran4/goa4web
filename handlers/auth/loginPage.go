package auth

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

type loginFormHandler struct{ msg string }

func (l loginFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, l.msg)
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
		rdh := handlers.RefreshDirectHandler{TargetURL: h.BackURL}
		cd.AutoRefresh = rdh.Content()
		handlers.TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", rdh)
		return
	}

	type Data struct {
		*common.CoreData
		BackURL string
		Method  string
		Values  url.Values
	}
	data := Data{
		CoreData: cd,
		BackURL:  h.BackURL,
		Method:   h.Method,
		Values:   h.Values,
	}
	if err := cd.ExecuteSiteTemplate(w, r, "redirectBackPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

var _ http.Handler = (*redirectBackPageHandler)(nil)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg string) {
	type Data struct {
		*common.CoreData
		Error   string
		Code    string
		Back    string
		BackSig string
		BackTS  string
		Method  string
		Data    string
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Error:    errMsg,
		Code:     r.FormValue("code"),
		Back:     r.Context().Value(consts.KeyCoreData).(*common.CoreData).SanitizeBackURL(r, r.FormValue("back")),
		BackSig:  r.FormValue("back_sig"),
		BackTS:   r.FormValue("back_ts"),
		Method:   r.FormValue("method"),
		Data:     r.FormValue("data"),
	}
	handlers.TemplateHandler(w, r, "loginPage.gohtml", data)
}
