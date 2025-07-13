package common

import "github.com/arran4/goa4web/internal/eventbus"

// NewTaskEvent returns a BasicTaskEvent for the given task name.
func NewTaskEvent(name string) eventbus.BasicTaskEvent {
	return eventbus.BasicTaskEvent{EventName: name, Match: TaskMatcher(name)}
}
