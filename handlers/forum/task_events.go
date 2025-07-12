package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// ReplyTask describes posting a reply to a forum thread.
var ReplyTask = eventbus.TaskEvent{
	Name:    hcommon.TaskReply,
	Matcher: hcommon.TaskMatcher(hcommon.TaskReply),
	Notification: func(path string, userID int32, data map[string]any) eventbus.EventNotification {
		return eventbus.EventNotification{
			Source:       hcommon.TaskReply,
			Path:         path,
			UserID:       userID,
			TemplateData: data,
		}
	},
}
