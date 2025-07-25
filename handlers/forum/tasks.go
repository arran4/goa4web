package forum

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskCreateThread creates a new forum thread.
	TaskCreateThread tasks.TaskString = "Create Thread"

	// TaskReply posts a reply to a thread.
	TaskReply tasks.TaskString = "Reply"

	// TaskEditReply edits a comment or reply.
	TaskEditReply tasks.TaskString = "Edit Reply"

	// TaskCancel cancels the current operation and returns to the previous page.
	TaskCancel tasks.TaskString = "Cancel"

	// TaskGrantRole grants a role to a forum user.
	TaskGrantRole = "Grant role"

	// TaskUpdateRole updates an existing forum role grant.
	TaskUpdateRole = "Update role"

	// TaskRevokeRole revokes a role from a forum user.
	TaskRevokeRole = "Revoke role"

	// TaskSetTopicRestriction adds a topic restriction.
	TaskSetTopicRestriction = "Set topic restriction"

	// TaskUpdateTopicRestriction updates a topic restriction.
	TaskUpdateTopicRestriction = "Update topic restriction"

	// TaskDeleteTopicRestriction deletes a topic restriction.
	TaskDeleteTopicRestriction = "Delete topic restriction"

	// TaskCopyTopicRestriction copies topic restrictions between topics.
	TaskCopyTopicRestriction = "Copy topic restriction"

	// TaskRemakeStatisticInformationOnForumthread refreshes thread statistics.
	TaskRemakeStatisticInformationOnForumthread = "Remake statistic information on forumthread"

	// TaskRemakeStatisticInformationOnForumtopic refreshes topic statistics.
	TaskRemakeStatisticInformationOnForumtopic = "Remake statistic information on forumtopic"

	// TaskForumCategoryChange updates a forum category name.
	TaskForumCategoryChange = "Forum category change"

	// TaskForumCategoryCreate creates a new forum category.
	TaskForumCategoryCreate = "Forum category create"

	// TaskDeleteCategory removes a forum category.
	TaskDeleteCategory = "Delete Category"

	// TaskForumThreadDelete removes a forum thread.
	TaskForumThreadDelete = "Forum thread delete"

	// TaskForumTopicChange updates a forum topic title.
	TaskForumTopicChange = "Forum topic change"

	// TaskForumTopicDelete removes a forum topic.
	TaskForumTopicDelete = "Forum topic delete"

	// TaskForumTopicCreate creates a new forum topic.
	TaskForumTopicCreate = "Forum topic create"

	// TaskTopicGrantCreate adds a new grant to a forum topic.
	TaskTopicGrantCreate tasks.TaskString = "Create grant"

	// TaskTopicGrantDelete removes an existing forum topic grant.
	TaskTopicGrantDelete tasks.TaskString = "Delete grant"

	// TaskSubscribeToTopic subscribes the user to new threads in a topic.
	TaskSubscribeToTopic tasks.TaskString = "Subscribe To Topic"

	// TaskUnsubscribeFromTopic removes topic thread notifications.
	TaskUnsubscribeFromTopic tasks.TaskString = "Unsubscribe From Topic"
)
