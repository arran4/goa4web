package forumcommon

import "github.com/arran4/goa4web/internal/tasks"

const (
	// TaskCreateThread creates a new forum thread.
	TaskCreateThread tasks.TaskString = "Create Thread"
	// TaskSubscribeToTopic subscribes the user to new threads in a topic.
	TaskSubscribeToTopic tasks.TaskString = "Subscribe To Topic"
	// TaskUnsubscribeFromTopic removes topic thread notifications.
	TaskUnsubscribeFromTopic tasks.TaskString = "Unsubscribe From Topic"
	// TaskMarkThreadRead marks a thread as read for the current user.
	TaskMarkThreadRead tasks.TaskString = "Mark Thread Read"
)
