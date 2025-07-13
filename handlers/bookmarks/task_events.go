package bookmarks

import hcommon "github.com/arran4/goa4web/handlers/common"

// SaveTask represents saving bookmark edits.
var SaveTask = hcommon.NewTaskEvent(hcommon.TaskSave)

// CreateTask represents creating a bookmark.
var CreateTask = hcommon.NewTaskEvent(hcommon.TaskCreate)
