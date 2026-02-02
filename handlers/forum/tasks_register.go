package forum

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers forum related tasks with the global registry.
func RegisterTasks() []tasks.NamedTask {
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
		subscribeTopicTaskAction,
		unsubscribeTopicTaskAction,
		addPublicLabelTask,
		removePublicLabelTask,
		addAuthorLabelTask,
		removeAuthorLabelTask,
		addPrivateLabelTask,
		removePrivateLabelTask,
		markThreadReadTask,
		setLabelsTask,
		addTopicPublicLabelTask,
		removeTopicPublicLabelTask,
		setTopicLabelsTask,
	}
}
