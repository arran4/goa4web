package images

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns image handling tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		uploadImageTask,
	}
}
