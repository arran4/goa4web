package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/tasks"
)

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskCreateThread creates a new forum thread.
	TaskCreateThread = forumcommon.TaskCreateThread

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

	// TaskTopicGrantUpdate updates grants for a forum topic action.
	TaskTopicGrantUpdate tasks.TaskString = "Update grants"

	// TaskCategoryGrantCreate adds a new grant to a forum category.
	TaskCategoryGrantCreate tasks.TaskString = "Create grant"

	// TaskCategoryGrantDelete removes an existing forum category grant.
	TaskCategoryGrantDelete tasks.TaskString = "Delete grant"

	// TaskSubscribeToTopic subscribes the user to new threads in a topic.
	TaskSubscribeToTopic = forumcommon.TaskSubscribeToTopic

	// TaskUnsubscribeFromTopic removes topic thread notifications.
	TaskUnsubscribeFromTopic = forumcommon.TaskUnsubscribeFromTopic

	// TaskAddPublicLabel adds a public label to a topic.
	TaskAddPublicLabel tasks.TaskString = "Add Public Label"

	// TaskRemovePublicLabel removes a public label from a topic.
	TaskRemovePublicLabel tasks.TaskString = "Remove Public Label"

	// TaskAddAuthorLabel adds an author-only label to a topic.
	TaskAddAuthorLabel tasks.TaskString = "Add Author Label"

	// TaskRemoveAuthorLabel removes an author-only label from a topic.
	TaskRemoveAuthorLabel tasks.TaskString = "Remove Author Label"

	// TaskAddPrivateLabel adds a private label to a topic.
	TaskAddPrivateLabel tasks.TaskString = "Add Private Label"

	// TaskRemovePrivateLabel removes a private label from a topic.
	TaskRemovePrivateLabel tasks.TaskString = "Remove Private Label"

	// TaskMarkThreadRead marks a thread as read for the current user.
	TaskMarkThreadRead = forumcommon.TaskMarkThreadRead

	// TaskSetLabels replaces public and private labels on a topic.
	TaskSetLabels tasks.TaskString = "Set Labels"
)
