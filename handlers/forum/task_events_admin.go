package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// SetUserLevelTask updates a user's forum access level.
type SetUserLevelTask struct{ tasks.TaskString }

var setUserLevelTask = &SetUserLevelTask{TaskString: TaskSetUserLevel}

// UpdateUserLevelTask modifies a user's access level.
type UpdateUserLevelTask struct{ tasks.TaskString }

var updateUserLevelTask = &UpdateUserLevelTask{TaskString: TaskUpdateUserLevel}

// DeleteUserLevelTask removes a user's access level.
type DeleteUserLevelTask struct{ tasks.TaskString }

var deleteUserLevelTask = &DeleteUserLevelTask{TaskString: TaskDeleteUserLevel}

// SetTopicRestrictionTask adds a topic restriction.
type SetTopicRestrictionTask struct{ tasks.TaskString }

var setTopicRestrictionTask = &SetTopicRestrictionTask{TaskString: TaskSetTopicRestriction}

// UpdateTopicRestrictionTask updates a topic restriction.
type UpdateTopicRestrictionTask struct{ tasks.TaskString }

var updateTopicRestrictionTask = &UpdateTopicRestrictionTask{TaskString: TaskUpdateTopicRestriction}

// DeleteTopicRestrictionTask deletes a topic restriction.
type DeleteTopicRestrictionTask struct{ tasks.TaskString }

var deleteTopicRestrictionTask = &DeleteTopicRestrictionTask{TaskString: TaskDeleteTopicRestriction}

// CopyTopicRestrictionTask copies topic restrictions between topics.
type CopyTopicRestrictionTask struct{ tasks.TaskString }

var copyTopicRestrictionTask = &CopyTopicRestrictionTask{TaskString: TaskCopyTopicRestriction}

// RemakeThreadStatsTask refreshes forum thread statistics.
type RemakeThreadStatsTask struct{ tasks.TaskString }

var remakeThreadStatsTask = &RemakeThreadStatsTask{TaskString: TaskRemakeStatisticInformationOnForumthread}

// RemakeTopicStatsTask refreshes forum topic statistics.
type RemakeTopicStatsTask struct{ tasks.TaskString }

var remakeTopicStatsTask = &RemakeTopicStatsTask{TaskString: TaskRemakeStatisticInformationOnForumtopic}

// CategoryChangeTask updates a forum category name.
type CategoryChangeTask struct{ tasks.TaskString }

var categoryChangeTask = &CategoryChangeTask{TaskString: TaskForumCategoryChange}

// CategoryCreateTask creates a new forum category.
type CategoryCreateTask struct{ tasks.TaskString }

var categoryCreateTask = &CategoryCreateTask{TaskString: TaskForumCategoryCreate}

// DeleteCategoryTask removes a forum category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}

// ThreadDeleteTask removes a forum thread.
type ThreadDeleteTask struct{ tasks.TaskString }

var threadDeleteTask = &ThreadDeleteTask{TaskString: TaskForumThreadDelete}

// TopicChangeTask updates a forum topic title.
type TopicChangeTask struct{ tasks.TaskString }

var topicChangeTask = &TopicChangeTask{TaskString: TaskForumTopicChange}

// TopicDeleteTask removes a forum topic.
type TopicDeleteTask struct{ tasks.TaskString }

var topicDeleteTask = &TopicDeleteTask{TaskString: TaskForumTopicDelete}

// TopicCreateTask creates a new forum topic.
type TopicCreateTask struct{ tasks.TaskString }

var topicCreateTask = &TopicCreateTask{TaskString: TaskForumTopicCreate}
