package linker

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns linker related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		AdminAddTask,
		UpdateCategoryTask,
		RenameCategoryTask,
		AdminDeleteCategoryTask,
		CreateCategoryTask,
		AdminDeleteTask,
		AdminApproveTask,
		AdminBulkDeleteTask,
		AdminBulkApproveTask,
		AdminEditLinkTask,
		UserAllowTask,
		UserDisallowTask,
		categoryGrantCreateTask,
		AdminCategoryGrantDeleteTask,
		commentEditAction,
		commentEditActionCancel,
		replyTaskEvent,
		suggestTask,
	}
}
