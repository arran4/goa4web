package bookmarks

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns bookmark related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		saveTask,
		createTask,
	}
}
