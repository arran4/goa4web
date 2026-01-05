package privateforum

import "github.com/arran4/goa4web/internal/tasks"

const (
	// TaskPrivateTopicCreate creates a new private conversation topic.
	TaskPrivateTopicCreate tasks.TaskString = "Private topic create"
)

const (
	PrivateForumStartDiscussionPageTmpl    = "privateforum/start_discussion.gohtml"
	PrivateForumCreateTopicPageTmpl        = "forum/create_topic.gohtml"
	PrivateForumTopicsOnlyTmpl             = "privateforum/topics_only.gohtml"
)
