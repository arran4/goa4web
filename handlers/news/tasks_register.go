package news

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers news related tasks with the global registry.
func RegisterTasks() {
	tasks.Register(NewsUserAllowTask)
	tasks.Register(NewsUserRemoveTask)
	tasks.Register(announcementAddTask)
	tasks.Register(announcementDeleteTask)
	tasks.Register(editReplyTask)
	tasks.Register(cancelTask)
	tasks.Register(replyTask)
	tasks.Register(editTask)
	tasks.Register(newPostTask)
}
