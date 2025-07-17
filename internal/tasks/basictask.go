package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// BasicTaskEvent describes an action accessible via the router.
type BasicTaskEvent struct {
	EventName     string
	Match         mux.MatcherFunc
	ActionHandler http.HandlerFunc
}

func (b BasicTaskEvent) Name() string { return b.EventName }

func (b BasicTaskEvent) Matcher() mux.MatcherFunc {
	if b.Match != nil {
		return b.Match
	}
	return HasTask(b.EventName)
}

func (b BasicTaskEvent) Action(w http.ResponseWriter, r *http.Request) {
	if b.ActionHandler != nil {
		b.ActionHandler(w, r)
	}
}

func (BasicTaskEvent) IsAdminTask() bool { return true }

// NewTaskEvent creates a BasicTaskEvent with the given name.
func NewTaskEvent(name string) BasicTaskEvent {
	return BasicTaskEvent{EventName: name, Match: HasTask(name)}
}

// NewTaskEventWithHandlers creates a BasicTaskEvent with custom handlers.
func NewTaskEventWithHandlers(name string, page, action http.HandlerFunc) BasicTaskEvent {
	_ = page // page handler currently unused
	return BasicTaskEvent{EventName: name, Match: HasTask(name), ActionHandler: action}
}
