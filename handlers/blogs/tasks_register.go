package blogs

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers blog related tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		addBlogTask,
		editBlogTask,
		replyBlogTask,
		editReplyTask,
		cancelTask,
		setLabelsTask,
	}
}
