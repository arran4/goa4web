package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// TODO refactor this out entirely

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
	return HasTask(b, b.EventName)
}

func (b BasicTaskEvent) Action(w http.ResponseWriter, r *http.Request) {
	if b.ActionHandler != nil {
		b.ActionHandler(w, r)
	}
}

func (BasicTaskEvent) IsAdminTask() bool { return true }

// NewTaskEvent creates a BasicTaskEvent with the given name.
func NewTaskEvent(name string) BasicTaskEvent {
	b := BasicTaskEvent{EventName: name}
	b.Match = HasTask(b, name)
	return b
}
