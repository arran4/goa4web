package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HasTask restricts requests to those specifying the provided task value.
func HasTask(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.Method == http.MethodPost {
			return r.PostFormValue("task") == taskName
		}
		return r.FormValue("task") == taskName
	}
}

// HasNoTask matches requests that do not specify a task.
func HasNoTask() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.Method == http.MethodPost {
			return r.PostFormValue("task") == ""
		}
		return r.FormValue("task") == ""
	}
}
