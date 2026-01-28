package news

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/tasks"
)

// Task constants used within the news handlers.
const (
	// TaskAdd represents adding an announcement.
	TaskAdd tasks.TaskString = "Add"
	// TaskDelete removes an item such as an announcement.
	TaskDelete = "Delete"
	// TaskEdit modifies an existing news post.
	TaskEdit = "Edit"
	// TaskNewPost creates a new news post.
	TaskNewPost tasks.TaskString = "New Post"
	// TaskReply posts a reply to a news thread.
	TaskReply tasks.TaskString = "Reply"
	// TaskEditReply edits a comment or reply.
	TaskEditReply tasks.TaskString = "Edit Reply"
	// TaskCancel cancels the current operation and returns to the previous page.
	TaskCancel tasks.TaskString = "Cancel"
	// TaskUserAllow grants a role to a user.
	TaskUserAllow = "User Allow"
	// TaskUserDisallow removes a role from a user.
	TaskUserDisallow = "User Disallow"

	// TaskMarkRead marks a news item as read for the current user.
	TaskMarkRead tasks.TaskString = forumcommon.TaskMarkThreadRead
)
