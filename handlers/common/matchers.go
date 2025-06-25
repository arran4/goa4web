package common

import (
	"net/http"

	"github.com/gorilla/mux"
)

// TaskMatcher restricts requests to those specifying the provided task value.
func TaskMatcher(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return r.PostFormValue("task") == taskName
	}
}

// NoTask matches requests that do not specify a task.
func NoTask() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return r.PostFormValue("task") == ""
	}
}
