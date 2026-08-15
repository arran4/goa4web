package forum

import (
	"fmt"
	"strings"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func privateThreadSubscriberGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if !strings.HasPrefix(evt.Path, "/private/topic/") {
		return nil, nil
	}
	data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData)
	if !ok || data.ThreadID == 0 {
		return nil, fmt.Errorf("private thread notification context not provided")
	}
	return []notif.GrantRequirement{{
		Section: consts.PermissionSectionPrivateForumThread,
		Item:    consts.PermissionItemThread,
		ItemID:  data.ThreadID,
		Action:  consts.PermissionActionView,
	}}, nil
}
