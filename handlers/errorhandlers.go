package handlers

import "net/http"

type errRedirectOnSamePageHandler struct {
	error
}

func (e errRedirectOnSamePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "?error="+e.Error(), http.StatusTemporaryRedirect)
}

var _ http.Handler = (*errRedirectOnSamePageHandler)(nil)

func ErrRedirectOnSamePageHandler(err error) error {
	return &errRedirectOnSamePageHandler{err}
}

type SessionFetchFail struct{}
