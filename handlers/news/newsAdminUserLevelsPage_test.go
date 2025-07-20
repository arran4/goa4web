package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestNewsUserLevelTasksTemplates(t *testing.T) {
	var allow newsUserAllowTask
	tpl := allow.AdminEmailTemplate()
	if *tpl != *notif.NewEmailTemplates("newsPermissionEmail") {
		t.Errorf("allow template mismatch: %#v", tpl)
	}
	nt := allow.AdminInternalNotificationTemplate()
	if *nt != notif.NotificationTemplateFilenameGenerator("news_permission") {
		t.Errorf("allow notification mismatch: %s", *nt)
	}

	var remove newsUserRemoveTask
	tpl = remove.AdminEmailTemplate()
	if *tpl != *notif.NewEmailTemplates("newsPermissionEmail") {
		t.Errorf("remove template mismatch: %#v", tpl)
	}
	nt = remove.AdminInternalNotificationTemplate()
	if *nt != notif.NotificationTemplateFilenameGenerator("news_permission") {
		t.Errorf("remove notification mismatch: %s", *nt)
	}
}
