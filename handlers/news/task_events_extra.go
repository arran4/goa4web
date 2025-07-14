package news

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ReplyTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskReply,
	Match:         hcommon.TaskMatcher(hcommon.TaskReply),
	ActionHandler: NewsPostReplyActionPage,
}

var EditTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskEdit,
	Match:         hcommon.TaskMatcher(hcommon.TaskEdit),
	ActionHandler: NewsPostEditActionPage,
}

var AnnouncementAddTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskAdd,
	Match:         hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionHandler: NewsAnnouncementActivateActionPage,
}

var AnnouncementDeleteTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: NewsAnnouncementDeactivateActionPage,
}

var UserAllowTask = eventbus.BasicTaskEvent{
	EventName:     "User Allow",
	Match:         hcommon.TaskMatcher("User Allow"),
	ActionHandler: NewsUsersPermissionsPermissionUserAllowPage,
}

var UserDisallowTask = eventbus.BasicTaskEvent{
	EventName:     "User Disallow",
	Match:         hcommon.TaskMatcher("User Disallow"),
	ActionHandler: NewsUsersPermissionsDisallowPage,
}
