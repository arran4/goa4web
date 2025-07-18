package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// BasicTaskEvent defines a simple Task implementation with optional matcher and action handler.
type BasicTaskEvent struct {
	TaskString
	Match         mux.MatcherFunc
	ActionHandler http.HandlerFunc
}

// Matcher returns the gorilla mux matcher for this task.
func (b BasicTaskEvent) Matcher() mux.MatcherFunc { return b.Match }

// Action executes the configured ActionHandler when set.
func (b BasicTaskEvent) Action(w http.ResponseWriter, r *http.Request) {
	if b.ActionHandler != nil {
		b.ActionHandler(w, r)
	}
}

// NewTaskEvent returns a BasicTaskEvent with the given task name.
func NewTaskEvent(name string) BasicTaskEvent {
	return BasicTaskEvent{TaskString: TaskString(name), Match: HasTask(name)}
}

// NewTaskEventWithHandlers returns a BasicTaskEvent using the provided page and action handlers.
func NewTaskEventWithHandlers(name string, page, action http.HandlerFunc) BasicTaskEvent {
	return BasicTaskEvent{TaskString: TaskString(name), Match: HasTask(name), ActionHandler: action}
}
