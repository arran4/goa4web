package forum

import "github.com/arran4/goa4web/internal/tasks"
import "github.com/arran4/goa4web/handlers/forumcommon"

// RemakeThreadStatsTask refreshes forum thread statistics.
type RemakeThreadStatsTask struct{ tasks.TaskString }

var remakeThreadStatsTask = &RemakeThreadStatsTask{TaskString: forumcommon.TaskRemakeStatisticInformationOnForumthread}

// ensure RemakeThreadStatsTask conforms to tasks.Task
var _ tasks.Task = (*RemakeThreadStatsTask)(nil)
