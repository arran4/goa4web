package common

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"
)

// NewTaskEvent returns a BasicTaskEvent for the given task name.
func NewTaskEvent(name string) eventbus.BasicTaskEvent {
	return eventbus.BasicTaskEvent{EventName: name, Match: TaskMatcher(name)}
}

// NewTaskEventWithHandlers creates a BasicTaskEvent and assigns the provided
// page and action handlers.
func NewTaskEventWithHandlers(name string, page, action http.HandlerFunc) eventbus.BasicTaskEvent {
	return eventbus.BasicTaskEvent{
		EventName: name,
		Match:     TaskMatcher(name),
		PageH:     page,
		ActionH:   action,
	}
}

var (
	// Generic task events used across multiple packages.
	AddTask       = NewTaskEvent(TaskAdd)
	CreateTask    = NewTaskEvent(TaskCreate)
	SaveTask      = NewTaskEvent(TaskSave)
	SaveAllTask   = NewTaskEvent(TaskSaveAll)
	DeleteTask    = NewTaskEvent(TaskDelete)
	CancelTask    = NewTaskEvent(TaskCancel)
	EditReplyTask = NewTaskEvent(TaskEditReply)
)
