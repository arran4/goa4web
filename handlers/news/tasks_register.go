package news

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers news related tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		announcementAddTask,
		announcementDeleteTask,
		editReplyTask,
		cancelTask,
		replyTask,
		editTask,
		deleteNewsPostTask,
		newPostTask,
		setLabelsTask,
	}
}
