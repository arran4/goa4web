package news

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var ReplyTask = tasks.BasicTaskEvent{
	EventName:     TaskReply,
	Match:         tasks.HasTask(TaskReply),
	ActionHandler: NewsPostReplyActionPage,
}

var EditTask = tasks.BasicTaskEvent{
	EventName:     TaskEdit,
	Match:         tasks.HasTask(TaskEdit),
	ActionHandler: NewsPostEditActionPage,
}

var AnnouncementAddTask = tasks.BasicTaskEvent{
	EventName:     TaskAdd,
	Match:         tasks.HasTask(TaskAdd),
	ActionHandler: NewsAnnouncementActivateActionPage,
}

var AnnouncementDeleteTask = tasks.BasicTaskEvent{
	EventName:     TaskDelete,
	Match:         tasks.HasTask(TaskDelete),
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
