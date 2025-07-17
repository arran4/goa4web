package news

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// NewPostTask represents creating a new news post.
var NewPostTask = tasks.BasicTaskEvent{
	EventName:     TaskNewPost,
	Match:         tasks.HasTask(TaskNewPost),
	ActionHandler: NewsPostNewActionPage,
}
