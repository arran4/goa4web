package linker

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ReplyTaskEvent = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskReply,
	Match:     hcommon.TaskMatcher(hcommon.TaskReply),
	ActionH:   CommentsReplyPage,
}

var EditReplyTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskEditReply,
	Match:     hcommon.TaskMatcher(hcommon.TaskEditReply),
	ActionH:   CommentEditActionPage,
}

var SuggestTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSuggest,
	Match:     hcommon.TaskMatcher(hcommon.TaskSuggest),
	ActionH:   SuggestActionPage,
}

var UpdateCategoryTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUpdate,
	Match:     hcommon.TaskMatcher(hcommon.TaskUpdate),
	ActionH:   AdminCategoriesUpdatePage,
}

var RenameCategoryTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskRenameCategory,
	Match:     hcommon.TaskMatcher(hcommon.TaskRenameCategory),
	ActionH:   AdminCategoriesRenamePage,
}

var DeleteCategoryTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDeleteCategory,
	Match:     hcommon.TaskMatcher(hcommon.TaskDeleteCategory),
	ActionH:   AdminCategoriesDeletePage,
}

var CreateCategoryTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskCreateCategory,
	Match:     hcommon.TaskMatcher(hcommon.TaskCreateCategory),
	ActionH:   AdminCategoriesCreatePage,
}

var AddTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskAdd,
	Match:     hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionH:   AdminAddActionPage,
}

var DeleteTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionH:   AdminQueueDeleteActionPage,
}

var ApproveTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskApprove,
	Match:     hcommon.TaskMatcher(hcommon.TaskApprove),
	ActionH:   AdminQueueApproveActionPage,
}

var BulkApproveTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskBulkApprove,
	Match:     hcommon.TaskMatcher(hcommon.TaskBulkApprove),
	ActionH:   AdminQueueBulkApproveActionPage,
}

var BulkDeleteTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskBulkDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskBulkDelete),
	ActionH:   AdminQueueBulkDeleteActionPage,
}

var UserAllowTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUserAllow,
	Match:     hcommon.TaskMatcher(hcommon.TaskUserAllow),
	ActionH:   AdminUserLevelsAllowActionPage,
}

var UserDisallowTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUserDisallow,
	Match:     hcommon.TaskMatcher(hcommon.TaskUserDisallow),
	ActionH:   AdminUserLevelsRemoveActionPage,
}
