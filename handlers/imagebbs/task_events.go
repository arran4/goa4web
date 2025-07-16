package imagebbs

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var UploadImageTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskUploadImage,
	Match:         tasks.HasTask(tasks.TaskUploadImage),
	ActionHandler: BoardPostImageActionPage,
}

var ReplyTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskReply,
	Match:         tasks.HasTask(tasks.TaskReply),
	ActionHandler: BoardThreadReplyActionPage,
}

var NewBoardTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskNewBoard,
	Match:         tasks.HasTask(tasks.TaskNewBoard),
	ActionHandler: AdminNewBoardMakePage,
}

var ModifyBoardTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskModifyBoard,
	Match:         tasks.HasTask(tasks.TaskModifyBoard),
	ActionHandler: AdminBoardModifyBoardActionPage,
}

var ApprovePostTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskApprove,
	Match:         tasks.HasTask(tasks.TaskApprove),
	ActionHandler: AdminApprovePostPage,
}
