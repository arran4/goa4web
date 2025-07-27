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

	// TaskWritingCategoryGrantCreate adds a new grant to a writing category.
	TaskWritingCategoryGrantCreate tasks.TaskString = "Create grant"

	// TaskWritingCategoryGrantDelete removes a grant from a writing category.
	TaskWritingCategoryGrantDelete tasks.TaskString = "Delete grant"
)
