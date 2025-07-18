package eventbus

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NamedTask exposes the name of a task.
// NamedTask exposes the name of a task.
type NamedTask interface{ TaskName() string }

// TaskEvent describes an application action that may trigger notifications.
// It provides handlers for performing the action and rendering the result as
// well as a factory for creating an EventNotification.
type TaskEvent interface {
	NamedTask
	Matcher() mux.MatcherFunc
	Action(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	BuildNotification(path string, userID int32, data map[string]any) EventNotification
}

// BasicTaskEvent is a simple implementation of the TaskEvent interface.
type BasicTaskEvent struct {
	// EventName is a short human readable name of the action.
	EventName string
	// Match restricts the route to requests specifying this task.
	Match mux.MatcherFunc
	// ActionHandler performs the task's logic.
	ActionHandler http.HandlerFunc
	// PageHandler is rendered once the task completes.
	PageHandler http.HandlerFunc
	// Notification builds an EventNotification for the executed task.
	Notification func(path string, userID int32, data map[string]any) EventNotification
}

// TaskName returns the task name.
func (e BasicTaskEvent) TaskName() string { return e.EventName }

// Name returns the task name.
func (e BasicTaskEvent) Name() string { return e.EventName }

// Matcher implements TaskEvent.
func (e BasicTaskEvent) Matcher() mux.MatcherFunc { return e.Match }

// Action implements TaskEvent.
func (e BasicTaskEvent) Action(w http.ResponseWriter, r *http.Request) {
	if e.ActionHandler != nil {
		e.ActionHandler(w, r)
	}
}

// Page implements TaskEvent.
func (e BasicTaskEvent) Page(w http.ResponseWriter, r *http.Request) {
	if e.PageHandler != nil {
		e.PageHandler(w, r)
	}
}

// BuildNotification creates a basic EventNotification when a custom builder is
// not supplied.
func (e BasicTaskEvent) BuildNotification(path string, userID int32, data map[string]any) EventNotification {
	if e.Notification != nil {
		return e.Notification(path, userID, data)
	}
	return EventNotification{
		Source:       e,
		Path:         path,
		UserID:       userID,
		TemplateData: data,
	}
}

var _ TaskEvent = BasicTaskEvent{}
