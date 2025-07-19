package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
)

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
