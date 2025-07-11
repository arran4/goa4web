package eventbus

import (
	"net/http"

	"github.com/gorilla/mux"
)

// TaskEvent describes an application action that may trigger notifications.
type TaskEvent struct {
	// Name is a short human readable name of the action.
	Name string
	// Matcher restricts the route to requests specifying this task.
	Matcher mux.MatcherFunc
	// Action performs the task's logic.
	Action http.HandlerFunc
	// Page is rendered once the task completes.
	Page http.HandlerFunc
	// Notification builds an EventNotification for the executed task.
	Notification func(path string, userID int32, data map[string]any) EventNotification
}

// BuildNotification creates a basic EventNotification when a custom builder is
// not supplied.
func (e TaskEvent) BuildNotification(path string, userID int32, data map[string]any) EventNotification {
	if e.Notification != nil {
		return e.Notification(path, userID, data)
	}
	return EventNotification{
		Source:       e.Name,
		Path:         path,
		UserID:       userID,
		TemplateData: data,
	}
}
