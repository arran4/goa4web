package news

import hcommon "github.com/arran4/goa4web/handlers/common"

var ReplyTask = hcommon.NewTaskEvent(hcommon.TaskReply)
var EditTask = hcommon.NewTaskEvent(hcommon.TaskEdit)
var AnnouncementAddTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var AnnouncementDeleteTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var UserAllowTask = hcommon.NewTaskEvent("User Allow")
var UserDisallowTask = hcommon.NewTaskEvent("User Disallow")
