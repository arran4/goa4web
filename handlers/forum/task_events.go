package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// ReplyTask describes posting a reply to a forum thread.
var ReplyTask = tasks.BasicTaskEvent{
	EventName:     TaskReply,
	Match:         tasks.HasTask(TaskReply),
	ActionHandler: TopicThreadReplyPage,
}

// CreateThreadTask describes creating a new forum thread.
var CreateThreadTask = tasks.BasicTaskEvent{
	EventName:     TaskCreateThread,
	Match:         tasks.HasTask(TaskCreateThread),
	PageHandler:   ThreadNewPage,
	ActionHandler: ThreadNewActionPage,
}
