package bookmarks

import hcommon "github.com/arran4/goa4web/handlers/common"

// SaveTask represents saving bookmark edits.
var SaveTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskSave, EditPage, EditSaveActionPage)

// CreateTask represents creating a bookmark.
var CreateTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskCreate, EditPage, EditCreateActionPage)
