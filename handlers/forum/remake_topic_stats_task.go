package forum

import "github.com/arran4/goa4web/internal/tasks"

// RemakeTopicStatsTask refreshes forum topic statistics.
type RemakeTopicStatsTask struct{ tasks.TaskString }

var remakeTopicStatsTask = &RemakeTopicStatsTask{TaskString: TaskRemakeStatisticInformationOnForumtopic}
