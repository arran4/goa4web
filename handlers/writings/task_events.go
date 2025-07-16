package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SubmitWritingTask represents submitting a new writing.
var SubmitWritingTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskSubmitWriting,
	Match:     tasks.HasTask(tasks.TaskSubmitWriting),
}

var ReplyTask = tasks.NewTaskEvent(tasks.TaskReply)
var EditReplyTask = tasks.NewTaskEvent(tasks.TaskEditReply)
var CancelTask = tasks.NewTaskEvent(tasks.TaskCancel)
var UpdateWritingTask = tasks.NewTaskEvent(tasks.TaskUpdateWriting)
