package imagebbs

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

var UploadImageTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUploadImage,
	Match:     hcommon.TaskMatcher(hcommon.TaskUploadImage),
	ActionH:   BoardPostImageActionPage,
}

var ReplyTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskReply,
	Match:     hcommon.TaskMatcher(hcommon.TaskReply),
	ActionH:   BoardThreadReplyActionPage,
}

var NewBoardTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskNewBoard,
	Match:     hcommon.TaskMatcher(hcommon.TaskNewBoard),
	ActionH:   AdminNewBoardMakePage,
}

var ModifyBoardTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskModifyBoard,
	Match:     hcommon.TaskMatcher(hcommon.TaskModifyBoard),
	ActionH:   AdminBoardModifyBoardActionPage,
}

var ApprovePostTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskApprove,
	Match:     hcommon.TaskMatcher(hcommon.TaskApprove),
	ActionH:   AdminApprovePostPage,
}
