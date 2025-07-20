package forum

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers forum related tasks with the global registry.
func RegisterTasks() {
	tasks.Register(createThreadTask)
	tasks.Register(replyTask)
}
