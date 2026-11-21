package forum

import (
	"fmt"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func privateThreadIDs(evt eventbus.TaskEvent) (topicID int32, threadID int32, private bool) {
	if evt.Data != nil {
		switch v := evt.Data[common.ForumReplyIntentEventKey].(type) {
		case common.ForumReplyNotificationIntent:
			topicID = v.TopicID
			threadID = v.ThreadID
			private = v.Private || strings.HasPrefix(v.CanonicalPath, "/private")
		case *common.ForumReplyNotificationIntent:
			if v != nil {
				topicID = v.TopicID
				threadID = v.ThreadID
				private = v.Private || strings.HasPrefix(v.CanonicalPath, "/private")
			}
		}
		if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
			if topicID == 0 {
				topicID = data.TopicID
			}
			if threadID == 0 {
				threadID = data.ThreadID
			}
		}
		if priv, ok := evt.Data["Private"].(bool); ok && priv {
			private = true
		}
	}
	if strings.HasPrefix(evt.Path, "/private/") || strings.HasPrefix(evt.Path, "/private/api/") {
		private = true
	}
	return topicID, threadID, private
}

func privateThreadSubscriberGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	topicID, threadID, private := privateThreadIDs(evt)
	if !private {
		return nil, nil
	}
	if topicID == 0 || threadID == 0 {
		return nil, fmt.Errorf("private thread notification context not provided")
	}
	return []notif.GrantRequirement{
		{
			Section: consts.PermissionSectionPrivateForum,
			Item:    consts.PermissionItemTopic,
			ItemID:  topicID,
			Action:  consts.PermissionActionView,
		},
		{
			Section: consts.PermissionSectionPrivateForumThread,
			Item:    consts.PermissionItemThread,
			ItemID:  threadID,
			Action:  consts.PermissionActionView,
		},
	}, nil
}
