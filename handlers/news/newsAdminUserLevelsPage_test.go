package news_test

import (
	"testing"

	admin "github.com/arran4/goa4web/handlers/admin"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestNewsUserLevelTasksTemplates(t *testing.T) {
	allow := admin.NewsUserAllowTask{TaskString: admin.TaskNewsUserAllow}
	tpl := allow.AdminEmailTemplate()
	if *tpl != *notif.NewEmailTemplates("newsPermissionEmail") {
		t.Errorf("allow template mismatch: %#v", tpl)
	}
	nt := allow.AdminInternalNotificationTemplate()
	if *nt != notif.NotificationTemplateFilenameGenerator("news_permission") {
		t.Errorf("allow notification mismatch: %s", *nt)
	}

	remove := admin.NewsUserRemoveTask{TaskString: admin.TaskNewsUserRemove}
	tpl = remove.AdminEmailTemplate()
	if *tpl != *notif.NewEmailTemplates("newsPermissionEmail") {
		t.Errorf("remove template mismatch: %#v", tpl)
	}
	nt = remove.AdminInternalNotificationTemplate()
	if *nt != notif.NotificationTemplateFilenameGenerator("news_permission") {
		t.Errorf("remove notification mismatch: %s", *nt)
	}
}
