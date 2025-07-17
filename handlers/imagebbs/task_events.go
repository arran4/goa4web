package imagebbs

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var UploadImageTask = tasks.BasicTaskEvent{
	EventName:     TaskUploadImage,
	Match:         tasks.HasTask(TaskUploadImage),
	ActionHandler: BoardPostImageActionPage,
}

var ReplyTask = tasks.BasicTaskEvent{
	EventName:     TaskReply,
	Match:         tasks.HasTask(TaskReply),
	ActionHandler: BoardThreadReplyActionPage,
}

var NewBoardTask = tasks.BasicTaskEvent{
	EventName:     TaskNewBoard,
	Match:         tasks.HasTask(TaskNewBoard),
	ActionHandler: AdminNewBoardMakePage,
}

var ModifyBoardTask = tasks.BasicTaskEvent{
	EventName:     TaskModifyBoard,
	Match:         tasks.HasTask(TaskModifyBoard),
	ActionHandler: AdminBoardModifyBoardActionPage,
}

var ApprovePostTask = tasks.BasicTaskEvent{
	EventName:     TaskApprove,
	Match:         tasks.HasTask(TaskApprove),
	ActionHandler: AdminApprovePostPage,
}
