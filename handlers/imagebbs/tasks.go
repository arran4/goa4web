package imagebbs

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskUploadImage uploads an image file to the image board.
	TaskUploadImage tasks.TaskString = "Upload image"

	// TaskReply posts a reply to a thread or comment.
	TaskReply tasks.TaskString = "Reply"

	// TaskNewBoard creates a new image board.
	TaskNewBoard tasks.TaskString = "New board"

	// TaskModifyBoard modifies the settings of an image board.
	TaskModifyBoard tasks.TaskString = "Modify board"

	// TaskApprove approves an item in moderation queues.
	TaskApprove tasks.TaskString = "Approve"

	// TaskModifyPost updates an existing image post.
	TaskModifyPost tasks.TaskString = "Modify image post"

	// TaskDeletePost removes an image post.
	TaskDeletePost tasks.TaskString = "Delete image post"
)
