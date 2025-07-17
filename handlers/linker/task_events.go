package linker

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var ReplyTaskEvent = tasks.BasicTaskEvent{
	EventName:     TaskReply,
	Match:         tasks.HasTask(TaskReply),
	ActionHandler: CommentsReplyPage,
}

var EditReplyTask = tasks.BasicTaskEvent{
	EventName:     TaskEditReply,
	Match:         tasks.HasTask(TaskEditReply),
	ActionHandler: CommentEditActionPage,
}

var SuggestTask = tasks.BasicTaskEvent{
	EventName:     TaskSuggest,
	Match:         tasks.HasTask(TaskSuggest),
	ActionHandler: SuggestActionPage,
}

var UpdateCategoryTask = tasks.BasicTaskEvent{
	EventName:     TaskUpdate,
	Match:         tasks.HasTask(TaskUpdate),
	ActionHandler: AdminCategoriesUpdatePage,
}

var RenameCategoryTask = tasks.BasicTaskEvent{
	EventName:     TaskRenameCategory,
	Match:         tasks.HasTask(TaskRenameCategory),
	ActionHandler: AdminCategoriesRenamePage,
}

var DeleteCategoryTask = tasks.BasicTaskEvent{
	EventName:     TaskDeleteCategory,
	Match:         tasks.HasTask(TaskDeleteCategory),
	ActionHandler: AdminCategoriesDeletePage,
}

var CreateCategoryTask = tasks.BasicTaskEvent{
	EventName:     TaskCreateCategory,
	Match:         tasks.HasTask(TaskCreateCategory),
	ActionHandler: AdminCategoriesCreatePage,
}

var AddTask = tasks.BasicTaskEvent{
	EventName:     TaskAdd,
	Match:         tasks.HasTask(TaskAdd),
	ActionHandler: AdminAddActionPage,
}

var DeleteTask = tasks.BasicTaskEvent{
	EventName:     TaskDelete,
	Match:         tasks.HasTask(TaskDelete),
	ActionHandler: AdminQueueDeleteActionPage,
}

var ApproveTask = tasks.BasicTaskEvent{
	EventName:     TaskApprove,
	Match:         tasks.HasTask(TaskApprove),
	ActionHandler: AdminQueueApproveActionPage,
}

var BulkApproveTask = tasks.BasicTaskEvent{
	EventName:     TaskBulkApprove,
	Match:         tasks.HasTask(TaskBulkApprove),
	ActionHandler: AdminQueueBulkApproveActionPage,
}

var BulkDeleteTask = tasks.BasicTaskEvent{
	EventName:     TaskBulkDelete,
	Match:         tasks.HasTask(TaskBulkDelete),
	ActionHandler: AdminQueueBulkDeleteActionPage,
}

var UserAllowTask = tasks.BasicTaskEvent{
	EventName:     TaskUserAllow,
	Match:         tasks.HasTask(TaskUserAllow),
	ActionHandler: AdminUserLevelsAllowActionPage,
}

var UserDisallowTask = tasks.BasicTaskEvent{
	EventName:     TaskUserDisallow,
	Match:         tasks.HasTask(TaskUserDisallow),
	ActionHandler: AdminUserLevelsRemoveActionPage,
}
