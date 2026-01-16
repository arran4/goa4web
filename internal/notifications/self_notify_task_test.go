package notifications

import (
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

type selfNotifyTaskForTest struct{ tasks.TaskString }

func (s selfNotifyTaskForTest) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool) {
	return nil, false
}

func (s selfNotifyTaskForTest) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	t := "notifications/password_reset.gotxt"
	return &t
}
