package linker

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action.
const (
	// TaskReply posts a reply to a thread or comment.
	TaskReply = "Reply"

	// TaskEditReply edits a comment or reply.
	TaskEditReply = "Edit Reply"

	// TaskSuggest creates a suggestion in the linker.
	TaskSuggest = "Suggest"

	// TaskUpdate updates an existing item.
	TaskUpdate = "Update"

	// TaskRenameCategory renames a category.
	TaskRenameCategory = "Rename Category"

	// TaskDeleteCategory removes a category.
	TaskDeleteCategory = "Delete Category"

	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory = "Create Category"

	// TaskAdd represents the "Add" action, commonly used when creating a new
	// record.
	TaskAdd = "Add"

	// TaskDelete removes an existing item.
	TaskDelete = "Delete"

	// TaskApprove approves an item in moderation queues.
	TaskApprove = "Approve"

	// TaskBulkApprove approves multiple queued items at once.
	TaskBulkApprove = "Bulk Approve"

	// TaskBulkDelete removes multiple queued items at once.
	TaskBulkDelete = "Bulk Delete"

	// TaskUserAllow grants a user a permission or level.
	TaskUserAllow = "User Allow"

	// TaskUserDisallow removes a user's permission or level.
	TaskUserDisallow = "User Disallow"
)
