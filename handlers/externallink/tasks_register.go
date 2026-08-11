package externallink

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns external link related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		reloadExternalLinkTask,
		prefetchExternalLinkTask,
	}
}
