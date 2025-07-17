package blogs

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var AddBlogTask = tasks.NewTaskEventWithHandlers(TaskAdd, BlogAddPage, BlogAddActionPage)
var ReplyBlogTask = tasks.NewTaskEventWithHandlers(TaskReply, nil, BlogReplyPostPage)
var EditBlogTask = tasks.NewTaskEventWithHandlers(TaskEdit, BlogEditPage, BlogEditActionPage)
var UserAllowTask = tasks.NewTaskEvent(TaskUserAllow)
var UserDisallowTask = tasks.NewTaskEvent(TaskUserDisallow)
var UsersAllowTask = tasks.NewTaskEvent(TaskUsersAllow)
var UsersDisallowTask = tasks.NewTaskEvent(TaskUsersDisallow)
