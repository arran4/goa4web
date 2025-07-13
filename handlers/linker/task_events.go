package linker

import hcommon "github.com/arran4/goa4web/handlers/common"

var ReplyTaskEvent = hcommon.NewTaskEvent(hcommon.TaskReply)
var EditReplyTask = hcommon.NewTaskEvent(hcommon.TaskEditReply)
var SuggestTask = hcommon.NewTaskEvent(hcommon.TaskSuggest)
var UpdateCategoryTask = hcommon.NewTaskEvent(hcommon.TaskUpdate)
var RenameCategoryTask = hcommon.NewTaskEvent(hcommon.TaskRenameCategory)
var DeleteCategoryTask = hcommon.NewTaskEvent(hcommon.TaskDeleteCategory)
var CreateCategoryTask = hcommon.NewTaskEvent(hcommon.TaskCreateCategory)
var AddTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var DeleteTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var ApproveTask = hcommon.NewTaskEvent(hcommon.TaskApprove)
var BulkApproveTask = hcommon.NewTaskEvent(hcommon.TaskBulkApprove)
var BulkDeleteTask = hcommon.NewTaskEvent(hcommon.TaskBulkDelete)
var UserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var UserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
