package writings

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns writing related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		submitWritingTask,
		replyTask,
		editReplyTask,
		cancelTask,
		updateWritingTask,
		writingCategoryChangeTask,
		writingCategoryCreateTask,
		categoryGrantCreateTask,
		categoryGrantDeleteTask,
		setLabelsTask,
	}
}
