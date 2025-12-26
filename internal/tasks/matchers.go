package tasks

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

// HasTask restricts requests to those specifying the provided task value.
func HasTask(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			return r.PostFormValue("task") == taskName
		}
		if r.FormValue("task") != taskName {
			return false
		}
		ref := r.Referer()
		if ref == "" {
			return false
		}
		refURL, err := url.Parse(ref)
		if err != nil {
			return false
		}
		if refURL.Host != r.Host {
			return false
		}
		return refURL.Query().Get("task") != taskName
	}
}

// HasNoTask matches requests that do not specify a task.
func HasNoTask() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			return r.PostFormValue("task") == ""
		}
		return r.FormValue("task") == ""
	}
}
