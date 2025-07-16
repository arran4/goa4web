package linker

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var ReplyTaskEvent = tasks.BasicTaskEvent{
	EventName:     tasks.TaskReply,
	Match:         tasks.HasTask(tasks.TaskReply),
	ActionHandler: CommentsReplyPage,
}

var EditReplyTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskEditReply,
	Match:         tasks.HasTask(tasks.TaskEditReply),
	ActionHandler: CommentEditActionPage,
}

var SuggestTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskSuggest,
	Match:         tasks.HasTask(tasks.TaskSuggest),
	ActionHandler: SuggestActionPage,
}

var UpdateCategoryTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskUpdate,
	Match:         tasks.HasTask(tasks.TaskUpdate),
	ActionHandler: AdminCategoriesUpdatePage,
}

var RenameCategoryTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskRenameCategory,
	Match:         tasks.HasTask(tasks.TaskRenameCategory),
	ActionHandler: AdminCategoriesRenamePage,
}

var DeleteCategoryTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskDeleteCategory,
	Match:         tasks.HasTask(tasks.TaskDeleteCategory),
	ActionHandler: AdminCategoriesDeletePage,
}

var CreateCategoryTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskCreateCategory,
	Match:         tasks.HasTask(tasks.TaskCreateCategory),
	ActionHandler: AdminCategoriesCreatePage,
}

var AddTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskAdd,
	Match:         tasks.HasTask(tasks.TaskAdd),
	ActionHandler: AdminAddActionPage,
}

var DeleteTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskDelete,
	Match:         tasks.HasTask(tasks.TaskDelete),
	ActionHandler: AdminQueueDeleteActionPage,
}

var ApproveTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskApprove,
	Match:         tasks.HasTask(tasks.TaskApprove),
	ActionHandler: AdminQueueApproveActionPage,
}

var BulkApproveTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskBulkApprove,
	Match:         tasks.HasTask(tasks.TaskBulkApprove),
	ActionHandler: AdminQueueBulkApproveActionPage,
}

var BulkDeleteTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskBulkDelete,
	Match:         tasks.HasTask(tasks.TaskBulkDelete),
	ActionHandler: AdminQueueBulkDeleteActionPage,
}

var UserAllowTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskUserAllow,
	Match:         tasks.HasTask(tasks.TaskUserAllow),
	ActionHandler: AdminUserLevelsAllowActionPage,
}

var UserDisallowTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskUserDisallow,
	Match:         tasks.HasTask(tasks.TaskUserDisallow),
	ActionHandler: AdminUserLevelsRemoveActionPage,
}
