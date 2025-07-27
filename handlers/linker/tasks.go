package linker

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskReply posts a reply to a thread or comment.
	TaskReply tasks.TaskString = "Reply"

	// TaskEditReply edits a comment or reply.
	TaskEditReply tasks.TaskString = "Edit Reply"

	// TaskCancel cancels the current operation and returns to the previous
	// page.
	TaskCancel tasks.TaskString = "Cancel"

	// TaskSuggest creates a suggestion in the linker.
	TaskSuggest tasks.TaskString = "Suggest"

	// TaskUpdate updates an existing item.
	TaskUpdate tasks.TaskString = "Update"

	// TaskRenameCategory renames a category.
	TaskRenameCategory tasks.TaskString = "Rename Category"

	// TaskDeleteCategory removes a category.
	TaskDeleteCategory tasks.TaskString = "Delete Category"

	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory tasks.TaskString = "Create Category"

	// TaskAdd represents the "Add" action, commonly used when creating a new
	// record.
	TaskAdd tasks.TaskString = "Add"

	// TaskDelete removes an existing item.
	TaskDelete tasks.TaskString = "Delete"

	// TaskApprove approves an item in moderation queues.
	TaskApprove tasks.TaskString = "Approve"

	// TaskBulkApprove approves multiple queued items at once.
	TaskBulkApprove tasks.TaskString = "Bulk Approve"

	// TaskBulkDelete removes multiple queued items at once.
	TaskBulkDelete tasks.TaskString = "Bulk Delete"

	// TaskUserAllow grants a user a role.
	TaskUserAllow tasks.TaskString = "User Allow"

	// TaskUserDisallow removes a user's role.
	TaskUserDisallow tasks.TaskString = "User Disallow"

	// TaskCategoryGrantCreate adds a new grant to a linker category.
	TaskCategoryGrantCreate tasks.TaskString = "Create grant"

	// TaskCategoryGrantDelete removes an existing linker category grant.
	TaskCategoryGrantDelete tasks.TaskString = "Delete grant"
)
