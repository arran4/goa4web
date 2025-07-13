package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// ReplyTask describes posting a reply to a forum thread.
var ReplyTask = eventbus.TaskEvent{
	Name:    hcommon.TaskReply,
	Matcher: hcommon.TaskMatcher(hcommon.TaskReply),
}

// CreateThreadTask describes creating a new forum thread.
var CreateThreadTask = eventbus.TaskEvent{
	Name:    hcommon.TaskCreateThread,
	Matcher: hcommon.TaskMatcher(hcommon.TaskCreateThread),
}
