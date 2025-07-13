package blogs

import hcommon "github.com/arran4/goa4web/handlers/common"

var AddBlogTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskAdd, BlogAddPage, BlogAddActionPage)
var ReplyBlogTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskReply, nil, BlogReplyPostPage)
var EditBlogTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskEdit, BlogEditPage, BlogEditActionPage)
var UserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var UserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
var UsersAllowTask = hcommon.NewTaskEvent(hcommon.TaskUsersAllow)
var UsersDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUsersDisallow)
