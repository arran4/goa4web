package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var UserAllowTask = tasks.NewTaskEvent(TaskUserAllow)
var UserDisallowTask = tasks.NewTaskEvent(TaskUserDisallow)
var AddApprovalTask = tasks.NewTaskEvent(TaskAddApproval)
var UpdateApprovalTask = tasks.NewTaskEvent(TaskUpdateUserApproval)
var DeleteApprovalTask = tasks.NewTaskEvent(TaskDeleteUserApproval)
var WritingCategoryChangeTask = tasks.NewTaskEvent(TaskWritingCategoryChange)
var WritingCategoryCreateTask = tasks.NewTaskEvent(TaskWritingCategoryCreate)
