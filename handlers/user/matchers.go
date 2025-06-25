package user

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/core"
)

// RequiresAnAccount ensures the requester has a valid session.
func RequiresAnAccount() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}

// TaskMatcher restricts requests to those specifying the provided task.
func TaskMatcher(taskName string) mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return r.PostFormValue("task") == taskName
	}
}
