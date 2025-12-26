package forum

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/forumcommon"
)

// topicSubscriptionPattern returns the subscription pattern for new threads in a topic.
func topicSubscriptionPattern(topicID int32) string {
	return forumcommon.TopicSubscriptionPattern(topicID)
}

// subscribedToTopic reports whether cd follows new threads in topicID.
func subscribedToTopic(cd *common.CoreData, topicID int32) bool {
	return forumcommon.SubscribedToTopic(cd, topicID)
}
