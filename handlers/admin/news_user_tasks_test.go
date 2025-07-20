package admin

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestNewsUserTasksTemplates(t *testing.T) {
	admins := []notif.AdminEmailTemplateProvider{
		NewsUserAllowTask,
		NewsUserRemoveTask,
	}
	for _, p := range admins {
		et := p.AdminEmailTemplate()
		if et == nil || et.Text == "" || et.HTML == "" || et.Subject == "" {
			t.Errorf("incomplete templates for %T", p)
		}
	}
}
