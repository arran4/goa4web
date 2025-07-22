package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Task describes an application action that may trigger notifications.
// It provides handlers for performing the action and rendering the result as
// well as a factory for creating an EventNotification.
type Task interface {
	Action(w http.ResponseWriter, r *http.Request)
}

type TaskMatcher interface {
	Matcher() mux.MatcherFunc
}

type Name interface {
	Name() string
}

// ActionResult represents a follow-up action to execute after a task completes.
type ActionResult interface {
	Action(w http.ResponseWriter, r *http.Request)
}

// ActionResultV2 is implemented by tasks that may return a follow-up ActionResult
// along with an error.
type ActionResultV2 interface {
	Action(w http.ResponseWriter, r *http.Request) (ActionResult, error)
}

type TaskString string

func (t TaskString) Name() string {
	return string(t)
}

func (t TaskString) Action(http.ResponseWriter, *http.Request) {}

func (t TaskString) Matcher() mux.MatcherFunc {
	return HasTask(string(t))
}

var _ TaskMatcher = (TaskString)("")
var _ Name = (TaskString)("")
var _ Task = (TaskString)("")
