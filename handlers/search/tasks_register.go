package search

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns search related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		searchForumTask,
		searchNewsTask,
		searchLinkerTask,
		searchBlogsTask,
		searchWritingsTask,
		remakeCommentsTask,
		remakeNewsTask,
		remakeBlogTask,
		remakeLinkerTask,
		remakeWritingTask,
		remakeImageTask,
	}
}
