package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var UserAllowTask = tasks.NewTaskEvent(tasks.TaskUserAllow)
var UserDisallowTask = tasks.NewTaskEvent(tasks.TaskUserDisallow)
var AddApprovalTask = tasks.NewTaskEvent(tasks.TaskAddApproval)
var UpdateApprovalTask = tasks.NewTaskEvent(tasks.TaskUpdateUserApproval)
var DeleteApprovalTask = tasks.NewTaskEvent(tasks.TaskDeleteUserApproval)
var WritingCategoryChangeTask = tasks.NewTaskEvent(tasks.TaskWritingCategoryChange)
var WritingCategoryCreateTask = tasks.NewTaskEvent(tasks.TaskWritingCategoryCreate)
