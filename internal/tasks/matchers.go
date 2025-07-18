package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HasTask restricts requests to those specifying the provided task value.
func HasTask(task Task, taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		if r.PostFormValue("task") != taskName {
			return false
		}
		// TODO implement something like this
		//if cd, ok := r.Context().Value("coreData").(*common.CoreData); ok {
		//	if evt := cd.Event(); evt != nil {
		//		evt.Task = task
		//	}
		//}
		return true
	}
}

// HasNoTask matches requests that do not specify a task.
func HasNoTask() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return r.PostFormValue("task") == ""
	}
}
