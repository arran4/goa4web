package templates

import "testing"

func TestNotificationOpenTemplateExists(t *testing.T) {
	if !TemplateExists("domains/user/notificationOpen.gohtml") {
		t.Fatalf("missing notification open template")
	}
}
