package imagebbs

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers image board administration tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		approvePostTask,
		modifyBoardTask,
		newBoardTask,
	}
}
