package writings

import hcommon "github.com/arran4/goa4web/handlers/common"

var UserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var UserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
var AddApprovalTask = hcommon.NewTaskEvent(hcommon.TaskAddApproval)
var UpdateApprovalTask = hcommon.NewTaskEvent(hcommon.TaskUpdateUserApproval)
var DeleteApprovalTask = hcommon.NewTaskEvent(hcommon.TaskDeleteUserApproval)
var WritingCategoryChangeTask = hcommon.NewTaskEvent(hcommon.TaskWritingCategoryChange)
var WritingCategoryCreateTask = hcommon.NewTaskEvent(hcommon.TaskWritingCategoryCreate)
