package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// SetUserLevelTask updates a user's forum access level.
var SetUserLevelTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSetUserLevel,
	Match:     hcommon.TaskMatcher(hcommon.TaskSetUserLevel),
}

// UpdateUserLevelTask modifies a user's access level.
var UpdateUserLevelTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUpdateUserLevel,
	Match:     hcommon.TaskMatcher(hcommon.TaskUpdateUserLevel),
}

// DeleteUserLevelTask removes a user's access level.
var DeleteUserLevelTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDeleteUserLevel,
	Match:     hcommon.TaskMatcher(hcommon.TaskDeleteUserLevel),
}

// SetTopicRestrictionTask adds a topic restriction.
var SetTopicRestrictionTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSetTopicRestriction,
	Match:     hcommon.TaskMatcher(hcommon.TaskSetTopicRestriction),
}

// UpdateTopicRestrictionTask updates a topic restriction.
var UpdateTopicRestrictionTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUpdateTopicRestriction,
	Match:     hcommon.TaskMatcher(hcommon.TaskUpdateTopicRestriction),
}

// DeleteTopicRestrictionTask deletes a topic restriction.
var DeleteTopicRestrictionTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDeleteTopicRestriction,
	Match:     hcommon.TaskMatcher(hcommon.TaskDeleteTopicRestriction),
}

// CopyTopicRestrictionTask copies topic restrictions between topics.
var CopyTopicRestrictionTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskCopyTopicRestriction,
	Match:     hcommon.TaskMatcher(hcommon.TaskCopyTopicRestriction),
}

// RemakeThreadStatsTask refreshes forum thread statistics.
var RemakeThreadStatsTask = hcommon.NewTaskEvent(hcommon.TaskRemakeStatisticInformationOnForumthread)

// RemakeTopicStatsTask refreshes forum topic statistics.
var RemakeTopicStatsTask = hcommon.NewTaskEvent(hcommon.TaskRemakeStatisticInformationOnForumtopic)

// CategoryChangeTask updates a forum category name.
var CategoryChangeTask = hcommon.NewTaskEvent(hcommon.TaskForumCategoryChange)

// CategoryCreateTask creates a new forum category.
var CategoryCreateTask = hcommon.NewTaskEvent(hcommon.TaskForumCategoryCreate)

// DeleteCategoryTask removes a forum category.
var DeleteCategoryTask = hcommon.NewTaskEvent(hcommon.TaskDeleteCategory)

// ThreadDeleteTask removes a forum thread.
var ThreadDeleteTask = hcommon.NewTaskEvent(hcommon.TaskForumThreadDelete)

// TopicChangeTask updates a forum topic title.
var TopicChangeTask = hcommon.NewTaskEvent(hcommon.TaskForumTopicChange)

// TopicDeleteTask removes a forum topic.
var TopicDeleteTask = hcommon.NewTaskEvent(hcommon.TaskForumTopicDelete)

// TopicCreateTask creates a new forum topic.
var TopicCreateTask = hcommon.NewTaskEvent(hcommon.TaskForumTopicCreate)
