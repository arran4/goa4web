package forum

import (
	"fmt"
	"strings"

	common "github.com/arran4/goa4web/core/common"
)

// topicSubscriptionPattern returns the subscription pattern for new threads in a topic.
func topicSubscriptionPattern(topicID int32) string {
	return fmt.Sprintf("%s:/forum/topic/%d/*", strings.ToLower(TaskCreateThread.Name()), topicID)
}

// subscribedToTopic reports whether cd follows new threads in topicID.
func subscribedToTopic(cd *common.CoreData, topicID int32) bool {
	if cd == nil || cd.UserID == 0 {
		return false
	}
	return cd.Subscribed(topicSubscriptionPattern(topicID))
}
