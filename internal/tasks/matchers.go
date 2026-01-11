package tasks

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

// HasTask restricts requests to those specifying the provided task value.
// It is an alias for HasFormTask to enforce strict body checks by default.
func HasTask(taskName string) mux.MatcherFunc {
	return HasFormTask(taskName)
}

// HasFormTask restricts requests to those specifying the provided task value in the form body.
// It strictly ignores query parameters.
func HasFormTask(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return r.PostFormValue("task") == taskName
	}
}

// HasQueryTask restricts requests to those specifying the provided task value in the query string.
// It enforces Referer checks to prevent CSRF and loops.
func HasQueryTask(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.FormValue("task") != taskName {
			return false
		}
		// Strict Referer Check
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

// HasFormOrQueryTask allows the task to be present in either the Form Body (for POST/PUT/PATCH)
// or the Query String (for GET, with strict Referer checks).
func HasFormOrQueryTask(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			return HasFormTask(taskName)(r, m)
		}
		return HasQueryTask(taskName)(r, m)
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
