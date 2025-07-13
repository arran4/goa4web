package imagebbs

import hcommon "github.com/arran4/goa4web/handlers/common"

var UploadImageTask = hcommon.NewTaskEvent(hcommon.TaskUploadImage)
var ReplyTask = hcommon.NewTaskEvent(hcommon.TaskReply)
var NewBoardTask = hcommon.NewTaskEvent(hcommon.TaskNewBoard)
var ModifyBoardTask = hcommon.NewTaskEvent(hcommon.TaskModifyBoard)
var ApprovePostTask = hcommon.NewTaskEvent(hcommon.TaskApprove)
