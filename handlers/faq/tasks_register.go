package faq

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns FAQ related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		askTask,
		removeQuestionTask,
		renameCategoryTask,
		deleteCategoryTask,
		createCategoryTask,
		editQuestionTask,
		deleteQuestionTask,
		createQuestionTask,
	}
}
