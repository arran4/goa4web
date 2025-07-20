package user

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers user-related tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		testMailTask,
	}
}
