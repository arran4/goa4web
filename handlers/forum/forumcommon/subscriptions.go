package forumcommon

import (
	"fmt"
	"strings"

	"github.com/arran4/goa4web/core/common"
)

// TopicSubscriptionPattern returns the subscription pattern for new threads in a topic.
func TopicSubscriptionPattern(topicID int32) string {
	return fmt.Sprintf("%s:/forum/topic/%d/*", strings.ToLower(TaskCreateThread.Name()), topicID)
}

// SubscribedToTopic reports whether cd follows new threads in topicID.
func SubscribedToTopic(cd *common.CoreData, topicID int32) bool {
	if cd == nil || cd.UserID == 0 {
		return false
	}
	return cd.Subscribed(TopicSubscriptionPattern(topicID), "internal")
}
