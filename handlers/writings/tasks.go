package writings

import "github.com/arran4/goa4web/internal/tasks"

// Task constants used by writings handlers.
const (
	// TaskSubmitWriting submits a new writing.
	TaskSubmitWriting tasks.TaskString = "Submit writing"

	// TaskReply posts a reply on a writing.
	TaskReply = "Reply"

	// TaskEditReply edits a comment or reply.
	TaskEditReply = "Edit Reply"

	// TaskCancel cancels a pending edit.
	TaskCancel = "Cancel"

	// TaskUpdateWriting updates an existing writing.
	TaskUpdateWriting = "Update writing"

	// TaskUserAllow grants a user a role.
	TaskUserAllow = "User Allow"

	// TaskUserDisallow removes a user's role.
	TaskUserDisallow = "User Disallow"

	// TaskWritingCategoryChange changes a writing category name.
	TaskWritingCategoryChange = "writing category change"

	// TaskWritingCategoryCreate creates a new writing category.
	TaskWritingCategoryCreate = "writing category create"

	// TaskCategoryGrantCreate adds a new grant to a writing category.
	TaskCategoryGrantCreate tasks.TaskString = "Create grant"

	// TaskCategoryGrantDelete removes an existing grant from a writing category.
	TaskCategoryGrantDelete tasks.TaskString = "Delete grant"
)
