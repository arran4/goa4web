package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// ReplyTask describes posting a reply to a forum thread.
var ReplyTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskReply,
	Match:     hcommon.TaskMatcher(hcommon.TaskReply),
	ActionH:   TopicThreadReplyPage,
}

// CreateThreadTask describes creating a new forum thread.
var CreateThreadTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskCreateThread,
	Match:     hcommon.TaskMatcher(hcommon.TaskCreateThread),
	PageH:     ThreadNewPage,
	ActionH:   ThreadNewActionPage,
}
