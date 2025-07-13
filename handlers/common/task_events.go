package common

import "github.com/arran4/goa4web/internal/eventbus"

// NewTaskEvent returns a basic TaskEvent for the given task name.
func NewTaskEvent(name string) eventbus.TaskEvent {
	return eventbus.TaskEvent{Name: name, Matcher: TaskMatcher(name)}
}
