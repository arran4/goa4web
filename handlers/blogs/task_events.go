package blogs

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var AddBlogTask = tasks.NewTaskEventWithHandlers(tasks.TaskAdd, BlogAddPage, BlogAddActionPage)
var ReplyBlogTask = tasks.NewTaskEventWithHandlers(tasks.TaskReply, nil, BlogReplyPostPage)
var EditBlogTask = tasks.NewTaskEventWithHandlers(tasks.TaskEdit, BlogEditPage, BlogEditActionPage)
var UserAllowTask = tasks.NewTaskEvent(tasks.TaskUserAllow)
var UserDisallowTask = tasks.NewTaskEvent(tasks.TaskUserDisallow)
var UsersAllowTask = tasks.NewTaskEvent(tasks.TaskUsersAllow)
var UsersDisallowTask = tasks.NewTaskEvent(tasks.TaskUsersDisallow)
