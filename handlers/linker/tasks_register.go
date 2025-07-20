package linker

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns linker related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		AddTask,
		UpdateCategoryTask,
		RenameCategoryTask,
		DeleteCategoryTask,
		CreateCategoryTask,
		DeleteTask,
		ApproveTask,
		BulkDeleteTask,
		BulkApproveTask,
		UserAllowTask,
		UserDisallowTask,
		commentEditAction,
		commentEditActionCancel,
		replyTaskEvent,
		suggestTask,
	}
}
