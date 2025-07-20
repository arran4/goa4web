package auth

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers all auth package tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		forgotPasswordTask,
	}
}
