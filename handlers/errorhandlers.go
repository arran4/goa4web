package handlers

import "net/http"

type errRedirectOnSamePageHandler struct {
	error
}

func (e errRedirectOnSamePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	RedirectToGet(w, r, "?error="+e.Error())
}

var _ http.Handler = (*errRedirectOnSamePageHandler)(nil)

func ErrRedirectOnSamePageHandler(err error) error {
	return &errRedirectOnSamePageHandler{err}
}

type SessionFetchFail struct{}
