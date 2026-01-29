package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
	"github.com/arran4/goa4web/internal/tasks"
)

// RegisterTasks registers forum related tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
	// Instantiate common forum context to get subscription tasks
	commonForum := forumcommon.New("forum", "/forum")

	return []tasks.NamedTask{
		createThreadTask,
		replyTask,
		topicCreateTask,
		topicChangeTask,
		topicDeleteTask,
		topicGrantCreateTask,
		topicGrantUpdateTask,
		topicGrantDeleteTask,
		categoryGrantCreateTask,
		categoryGrantDeleteTask,
		commonForum.SubscribeTopicTask(),
		commonForum.UnsubscribeTopicTask(),
		addPublicLabelTask,
		removePublicLabelTask,
		addAuthorLabelTask,
		removeAuthorLabelTask,
		addPrivateLabelTask,
		removePrivateLabelTask,
		markThreadReadTask,
		setLabelsTask,
	}
}
