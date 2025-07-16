package news

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var ReplyTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskReply,
	Match:         tasks.HasTask(tasks.TaskReply),
	ActionHandler: NewsPostReplyActionPage,
}

var EditTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskEdit,
	Match:         tasks.HasTask(tasks.TaskEdit),
	ActionHandler: NewsPostEditActionPage,
}

var AnnouncementAddTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskAdd,
	Match:         tasks.HasTask(tasks.TaskAdd),
	ActionHandler: NewsAnnouncementActivateActionPage,
}

var AnnouncementDeleteTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskDelete,
	Match:         tasks.HasTask(tasks.TaskDelete),
	ActionHandler: NewsAnnouncementDeactivateActionPage,
}

var UserAllowTask = tasks.BasicTaskEvent{
	EventName:     "User Allow",
	Match:         tasks.HasTask("User Allow"),
	ActionHandler: NewsUsersPermissionsPermissionUserAllowPage,
}

var UserDisallowTask = tasks.BasicTaskEvent{
	EventName:     "User Disallow",
	Match:         tasks.HasTask("User Disallow"),
	ActionHandler: NewsUsersPermissionsDisallowPage,
}
