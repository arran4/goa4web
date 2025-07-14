package writings

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// SubmitWritingTask represents submitting a new writing.
var SubmitWritingTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSubmitWriting,
	Match:     hcommon.TaskMatcher(hcommon.TaskSubmitWriting),
}

var ReplyTask = hcommon.NewTaskEvent(hcommon.TaskReply)
var EditReplyTask = hcommon.NewTaskEvent(hcommon.TaskEditReply)
var CancelTask = hcommon.NewTaskEvent(hcommon.TaskCancel)
var UpdateWritingTask = hcommon.NewTaskEvent(hcommon.TaskUpdateWriting)
