package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SetUserLevelTask updates a user's forum access level.
var SetUserLevelTask = tasks.BasicTaskEvent{
	EventName: TaskSetUserLevel,
	Match:     tasks.HasTask(TaskSetUserLevel),
}

// UpdateUserLevelTask modifies a user's access level.
var UpdateUserLevelTask = tasks.BasicTaskEvent{
	EventName: TaskUpdateUserLevel,
	Match:     tasks.HasTask(TaskUpdateUserLevel),
}

// DeleteUserLevelTask removes a user's access level.
var DeleteUserLevelTask = tasks.BasicTaskEvent{
	EventName: TaskDeleteUserLevel,
	Match:     tasks.HasTask(TaskDeleteUserLevel),
}

// SetTopicRestrictionTask adds a topic restriction.
var SetTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: TaskSetTopicRestriction,
	Match:     tasks.HasTask(TaskSetTopicRestriction),
}

// UpdateTopicRestrictionTask updates a topic restriction.
var UpdateTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: TaskUpdateTopicRestriction,
	Match:     tasks.HasTask(TaskUpdateTopicRestriction),
}

// DeleteTopicRestrictionTask deletes a topic restriction.
var DeleteTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: TaskDeleteTopicRestriction,
	Match:     tasks.HasTask(TaskDeleteTopicRestriction),
}

// CopyTopicRestrictionTask copies topic restrictions between topics.
var CopyTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: TaskCopyTopicRestriction,
	Match:     tasks.HasTask(TaskCopyTopicRestriction),
}

// RemakeThreadStatsTask refreshes forum thread statistics.
var RemakeThreadStatsTask = tasks.NewTaskEvent(TaskRemakeStatisticInformationOnForumthread)

// RemakeTopicStatsTask refreshes forum topic statistics.
var RemakeTopicStatsTask = tasks.NewTaskEvent(TaskRemakeStatisticInformationOnForumtopic)

// CategoryChangeTask updates a forum category name.
var CategoryChangeTask = tasks.NewTaskEvent(TaskForumCategoryChange)

// CategoryCreateTask creates a new forum category.
var CategoryCreateTask = tasks.NewTaskEvent(TaskForumCategoryCreate)

// DeleteCategoryTask removes a forum category.
var DeleteCategoryTask = tasks.NewTaskEvent(TaskDeleteCategory)

// ThreadDeleteTask removes a forum thread.
var ThreadDeleteTask = tasks.NewTaskEvent(TaskForumThreadDelete)

// TopicChangeTask updates a forum topic title.
var TopicChangeTask = tasks.NewTaskEvent(TaskForumTopicChange)

// TopicDeleteTask removes a forum topic.
var TopicDeleteTask = tasks.NewTaskEvent(TaskForumTopicDelete)

// TopicCreateTask creates a new forum topic.
var TopicCreateTask = tasks.NewTaskEvent(TaskForumTopicCreate)
