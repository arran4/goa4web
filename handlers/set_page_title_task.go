package handlers

import "github.com/arran4/goa4web/internal/tasks"

// TaskMigratePageTitles updates handlers to use CoreData.SetPageTitle.
const TaskMigratePageTitles tasks.TaskString = "Migrate page titles"

// MigratePageTitlesTask is a placeholder task for migrating page titles to
// CoreData. The action implementation will be added in the future.
type MigratePageTitlesTask struct{ tasks.TaskString }

var migratePageTitlesTask = &MigratePageTitlesTask{TaskString: TaskMigratePageTitles}

// compile-time check ensuring MigratePageTitlesTask implements tasks.Task.
var _ tasks.Task = (*MigratePageTitlesTask)(nil)
