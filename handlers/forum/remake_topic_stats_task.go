package forum

import "github.com/arran4/goa4web/internal/tasks"
import "github.com/arran4/goa4web/handlers/forum/forumcommon"

// RemakeTopicStatsTask refreshes forum topic statistics.
type RemakeTopicStatsTask struct{ tasks.TaskString }

var remakeTopicStatsTask = &RemakeTopicStatsTask{TaskString: forumcommon.TaskRemakeStatisticInformationOnForumtopic}

// ensure RemakeTopicStatsTask conforms to tasks.Task
var _ tasks.Task = (*RemakeTopicStatsTask)(nil)
