package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SetUserLevelTask updates a user's forum access level.
var SetUserLevelTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskSetUserLevel,
	Match:     tasks.HasTask(tasks.TaskSetUserLevel),
}

// UpdateUserLevelTask modifies a user's access level.
var UpdateUserLevelTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskUpdateUserLevel,
	Match:     tasks.HasTask(tasks.TaskUpdateUserLevel),
}

// DeleteUserLevelTask removes a user's access level.
var DeleteUserLevelTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskDeleteUserLevel,
	Match:     tasks.HasTask(tasks.TaskDeleteUserLevel),
}

// SetTopicRestrictionTask adds a topic restriction.
var SetTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskSetTopicRestriction,
	Match:     tasks.HasTask(tasks.TaskSetTopicRestriction),
}

// UpdateTopicRestrictionTask updates a topic restriction.
var UpdateTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskUpdateTopicRestriction,
	Match:     tasks.HasTask(tasks.TaskUpdateTopicRestriction),
}

// DeleteTopicRestrictionTask deletes a topic restriction.
var DeleteTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskDeleteTopicRestriction,
	Match:     tasks.HasTask(tasks.TaskDeleteTopicRestriction),
}

// CopyTopicRestrictionTask copies topic restrictions between topics.
var CopyTopicRestrictionTask = tasks.BasicTaskEvent{
	EventName: tasks.TaskCopyTopicRestriction,
	Match:     tasks.HasTask(tasks.TaskCopyTopicRestriction),
}

// RemakeThreadStatsTask refreshes forum thread statistics.
var RemakeThreadStatsTask = tasks.NewTaskEvent(tasks.TaskRemakeStatisticInformationOnForumthread)

// RemakeTopicStatsTask refreshes forum topic statistics.
var RemakeTopicStatsTask = tasks.NewTaskEvent(tasks.TaskRemakeStatisticInformationOnForumtopic)

// CategoryChangeTask updates a forum category name.
var CategoryChangeTask = tasks.NewTaskEvent(tasks.TaskForumCategoryChange)

// CategoryCreateTask creates a new forum category.
var CategoryCreateTask = tasks.NewTaskEvent(tasks.TaskForumCategoryCreate)

// DeleteCategoryTask removes a forum category.
var DeleteCategoryTask = tasks.NewTaskEvent(tasks.TaskDeleteCategory)

// ThreadDeleteTask removes a forum thread.
var ThreadDeleteTask = tasks.NewTaskEvent(tasks.TaskForumThreadDelete)

// TopicChangeTask updates a forum topic title.
var TopicChangeTask = tasks.NewTaskEvent(tasks.TaskForumTopicChange)

// TopicDeleteTask removes a forum topic.
var TopicDeleteTask = tasks.NewTaskEvent(tasks.TaskForumTopicDelete)

// TopicCreateTask creates a new forum topic.
var TopicCreateTask = tasks.NewTaskEvent(tasks.TaskForumTopicCreate)
