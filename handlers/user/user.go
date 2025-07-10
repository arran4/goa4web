package user

import (
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

func redirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if session != nil {
		backURL := r.URL.RequestURI()
		session.Values["BackURL"] = backURL
		if r.Method != http.MethodGet {
			if err := r.ParseForm(); err == nil {
				session.Values["BackMethod"] = r.Method
				session.Values["BackData"] = r.Form.Encode()
			}
		} else {
			delete(session.Values, "BackMethod")
			delete(session.Values, "BackData")
		}
		_ = session.Save(r, w)
	}
	http.Redirect(w, r, "/login?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusTemporaryRedirect)
}
