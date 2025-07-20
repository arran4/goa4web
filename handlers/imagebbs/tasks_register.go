package imagebbs

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers image board administration tasks with the global registry.
func RegisterTasks() {
	tasks.Register(approvePostTask)
	tasks.Register(modifyBoardTask)
	tasks.Register(newBoardTask)
}
