package bookmarks

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveTask represents saving bookmark edits.
var SaveTask = tasks.NewTaskEventWithHandlers(TaskSave, EditPage, EditSaveActionPage)

// CreateTask represents creating a bookmark.
var CreateTask = tasks.NewTaskEventWithHandlers(TaskCreate, EditPage, EditCreateActionPage)
