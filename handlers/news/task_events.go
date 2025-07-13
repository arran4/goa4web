package news

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// NewPostTask represents creating a new news post.
var NewPostTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskNewPost,
	Match:     hcommon.TaskMatcher(hcommon.TaskNewPost),
	ActionH:   NewsPostNewActionPage,
}
