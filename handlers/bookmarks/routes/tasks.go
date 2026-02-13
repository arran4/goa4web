package routes

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action.
const (
	// TaskSave persists changes for a bookmark list.
	TaskSave tasks.TaskString = "Save"

	// TaskCreate creates a new bookmark list for the user.
	TaskCreate tasks.TaskString = "Create"
)
