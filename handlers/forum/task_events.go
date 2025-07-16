package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// ReplyTask describes posting a reply to a forum thread.
var ReplyTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskReply,
	Match:         tasks.HasTask(tasks.TaskReply),
	ActionHandler: TopicThreadReplyPage,
}

// CreateThreadTask describes creating a new forum thread.
var CreateThreadTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskCreateThread,
	Match:         tasks.HasTask(tasks.TaskCreateThread),
	PageHandler:   ThreadNewPage,
	ActionHandler: ThreadNewActionPage,
}
