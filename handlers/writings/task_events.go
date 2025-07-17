package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SubmitWritingTask represents submitting a new writing.
var SubmitWritingTask = tasks.BasicTaskEvent{
	EventName: TaskSubmitWriting,
	Match:     tasks.HasTask(TaskSubmitWriting),
}

var ReplyTask = tasks.NewTaskEvent(TaskReply)
var EditReplyTask = tasks.NewTaskEvent(TaskEditReply)
var CancelTask = tasks.NewTaskEvent(TaskCancel)
var UpdateWritingTask = tasks.NewTaskEvent(TaskUpdateWriting)
