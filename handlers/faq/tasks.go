package faq

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskAsk submits a new question to the FAQ system.
	TaskAsk = "Ask"

	// TaskAnswer submits an answer in the FAQ admin interface.
	TaskAnswer = "Answer"

	// TaskRemoveRemove removes an item, typically from a list.
	TaskRemoveRemove = "Remove"

	// TaskRenameCategory renames a category.
	TaskRenameCategory = "Rename Category"

	// TaskDeleteCategory removes a category.
	TaskDeleteCategory = "Delete Category"

	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory = "Create Category"

	// TaskEdit modifies an existing item.
	TaskEdit = "Edit"

	// TaskCreate indicates creation of an object.
	TaskCreate = "Create"

	// TaskCreateFromTemplate creates an FAQ entry from a template.
	TaskCreateFromTemplate = "Create from Template"

	// TaskAddCategoryGrant adds a grant to a category.
	TaskAddCategoryGrant = "Add Category Grant"

	// TaskRemoveCategoryGrant removes a grant from a category.
	TaskRemoveCategoryGrant = "Remove Category Grant"
)
