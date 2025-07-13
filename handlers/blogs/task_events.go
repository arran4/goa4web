package blogs

import hcommon "github.com/arran4/goa4web/handlers/common"

var AddBlogTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var ReplyBlogTask = hcommon.NewTaskEvent(hcommon.TaskReply)
var EditBlogTask = hcommon.NewTaskEvent(hcommon.TaskEdit)
var UserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var UserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
var UsersAllowTask = hcommon.NewTaskEvent(hcommon.TaskUsersAllow)
var UsersDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUsersDisallow)
