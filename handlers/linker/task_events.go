package linker

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ReplyTaskEvent = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskReply,
	Match:         hcommon.TaskMatcher(hcommon.TaskReply),
	ActionHandler: CommentsReplyPage,
}

var EditReplyTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskEditReply,
	Match:         hcommon.TaskMatcher(hcommon.TaskEditReply),
	ActionHandler: CommentEditActionPage,
}

var SuggestTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSuggest,
	Match:         hcommon.TaskMatcher(hcommon.TaskSuggest),
	ActionHandler: SuggestActionPage,
}

var UpdateCategoryTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskUpdate,
	Match:         hcommon.TaskMatcher(hcommon.TaskUpdate),
	ActionHandler: AdminCategoriesUpdatePage,
}

var RenameCategoryTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskRenameCategory,
	Match:         hcommon.TaskMatcher(hcommon.TaskRenameCategory),
	ActionHandler: AdminCategoriesRenamePage,
}

var DeleteCategoryTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDeleteCategory,
	Match:         hcommon.TaskMatcher(hcommon.TaskDeleteCategory),
	ActionHandler: AdminCategoriesDeletePage,
}

var CreateCategoryTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskCreateCategory,
	Match:         hcommon.TaskMatcher(hcommon.TaskCreateCategory),
	ActionHandler: AdminCategoriesCreatePage,
}

var AddTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskAdd,
	Match:         hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionHandler: AdminAddActionPage,
}

var DeleteTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: AdminQueueDeleteActionPage,
}

var ApproveTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskApprove,
	Match:         hcommon.TaskMatcher(hcommon.TaskApprove),
	ActionHandler: AdminQueueApproveActionPage,
}

var BulkApproveTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskBulkApprove,
	Match:         hcommon.TaskMatcher(hcommon.TaskBulkApprove),
	ActionHandler: AdminQueueBulkApproveActionPage,
}

var BulkDeleteTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskBulkDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskBulkDelete),
	ActionHandler: AdminQueueBulkDeleteActionPage,
}

var UserAllowTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskUserAllow,
	Match:         hcommon.TaskMatcher(hcommon.TaskUserAllow),
	ActionHandler: AdminUserLevelsAllowActionPage,
}

var UserDisallowTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskUserDisallow,
	Match:         hcommon.TaskMatcher(hcommon.TaskUserDisallow),
	ActionHandler: AdminUserLevelsRemoveActionPage,
}
