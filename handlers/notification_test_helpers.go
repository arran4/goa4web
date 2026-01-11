package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// NotificationTemplateTest describes a test case for notification template rendering.
type NotificationTemplateTest struct {
	Name           string
	Task           tasks.Task
	Event          eventbus.TaskEvent
	ExpectedEmails []string // Expected email template prefixes
	ExpectedNotifs []string // Expected notification template names
}

// CreateTestEvent is a helper to create a realistic event for testing.
func CreateTestEvent(task tasks.Task, path string, userID int32, data map[string]any) eventbus.TaskEvent {
	if data == nil {
		data = make(map[string]any)
	}

	// Add common fields if not present
	if _, ok := data["Username"]; !ok {
		data["Username"] = "testuser"
	}

	return eventbus.TaskEvent{
		Task:    task,
		Path:    path,
		UserID:  userID,
		Data:    data,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
}

// CreateTestCoreData creates a minimal CoreData for testing handlers.
func CreateTestCoreData() *common.CoreData {
	cfg := config.NewRuntimeConfig()
	cfg.EmailFrom = "test@example.com"

	cd := &common.CoreData{
		Config: cfg,
	}

	return cd
}

// CreateTestRequest creates an HTTP request with CoreData in context.
func CreateTestRequest(cd *common.CoreData) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	return req.WithContext(ctx)
}

// MockQuerier provides a minimal database implementation for template testing.
type MockQuerier struct {
	db.QuerierStub
}

// getEmailFuncs returns template functions for email rendering with event data.
func getEmailFuncs(evt eventbus.TaskEvent) map[string]any {
	// Provide basic template functions that might be needed
	return map[string]any{
		"Username": func() string {
			if u, ok := evt.Data["Username"].(string); ok {
				return u
			}
			return "Unknown"
		},
	}
}

// getNotificationFuncs returns template functions for notification rendering.
func getNotificationFuncs(evt eventbus.TaskEvent) map[string]any {
	return map[string]any{
		"Username": func() string {
			if u, ok := evt.Data["Username"].(string); ok {
				return u
			}
			return "Unknown"
		},
	}
}

// TestNotificationTemplates validates all notification templates for a set of tasks.
// This helper function can be called from any package's test files.
func TestNotificationTemplates(t TestingT, tests []NotificationTemplateTest) {
	for _, tc := range tests {
		evt := tc.Event
		task := tc.Task

		// Test admin email templates
		if provider, ok := task.(notif.AdminEmailTemplateProvider); ok {
			et, send := provider.AdminEmailTemplate(evt)
			if send && et != nil {
				RequireEmailTemplates(t, et, evt)
			}
			RequireNotificationTemplate(t, provider.AdminInternalNotificationTemplate(evt), evt)
		}

		// Test self notification templates
		if provider, ok := task.(notif.SelfNotificationTemplateProvider); ok {
			et, send := provider.SelfEmailTemplate(evt)
			if send && et != nil {
				RequireEmailTemplates(t, et, evt)
			}
			RequireNotificationTemplate(t, provider.SelfInternalNotificationTemplate(evt), evt)
		}

		// Test target user templates
		if provider, ok := task.(notif.TargetUsersNotificationProvider); ok {
			et, send := provider.TargetEmailTemplate(evt)
			if send && et != nil {
				RequireEmailTemplates(t, et, evt)
			}
			RequireNotificationTemplate(t, provider.TargetInternalNotificationTemplate(evt), evt)
		}

		// Test subscriber templates
		if provider, ok := task.(notif.SubscribersNotificationTemplateProvider); ok {
			et, send := provider.SubscribedEmailTemplate(evt)
			if send && et != nil {
				RequireEmailTemplates(t, et, evt)
			}
			RequireNotificationTemplate(t, provider.SubscribedInternalNotificationTemplate(evt), evt)
		}

		// Test direct email templates
		if provider, ok := task.(notif.DirectEmailNotificationTemplateProvider); ok {
			et, send := provider.DirectEmailTemplate(evt)
			if send && et != nil {
				RequireEmailTemplates(t, et, evt)
			}
		}
	}
}

// RequireEmailTemplates validates that email templates exist and can be rendered with event data.
func RequireEmailTemplates(t TestingT, et *notif.EmailTemplates, evt eventbus.TaskEvent) {
	t.Helper()

	// Get template functions with event data
	funcs := getEmailFuncs(evt)

	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(funcs)
	textTmpls := templates.GetCompiledEmailTextTemplates(funcs)

	// Check HTML template exists
	if et.HTML != "" {
		tmpl := htmlTmpls.Lookup(et.HTML)
		if tmpl == nil {
			t.Errorf("missing html template %s", et.HTML)
		} else {
			// Try to render with event data
			if err := tmpl.Execute(httptest.NewRecorder(), evt.Data); err != nil {
				t.Errorf("html template %s render error: %v", et.HTML, err)
			}
		}
	}

	// Check text template exists
	if et.Text != "" {
		tmpl := textTmpls.Lookup(et.Text)
		if tmpl == nil {
			t.Errorf("missing text template %s", et.Text)
		} else {
			// Try to render with event data
			if err := tmpl.Execute(httptest.NewRecorder(), evt.Data); err != nil {
				t.Errorf("text template %s render error: %v", et.Text, err)
			}
		}
	}

	// Check subject template exists
	if et.Subject != "" {
		tmpl := textTmpls.Lookup(et.Subject)
		if tmpl == nil {
			t.Errorf("missing subject template %s", et.Subject)
		} else {
			// Try to render with event data
			if err := tmpl.Execute(httptest.NewRecorder(), evt.Data); err != nil {
				t.Errorf("subject template %s render error: %v", et.Subject, err)
			}
		}
	}
}

// RequireNotificationTemplate validates that internal notification templates exist and can be rendered.
func RequireNotificationTemplate(t TestingT, name *string, evt eventbus.TaskEvent) {
	t.Helper()

	if name == nil {
		return
	}

	funcs := getNotificationFuncs(evt)
	tmpl := templates.GetCompiledNotificationTemplates(funcs)

	notifTmpl := tmpl.Lookup(*name)
	if notifTmpl == nil {
		t.Errorf("missing notification template %s", *name)
	} else {
		// Try to render with event data
		if err := notifTmpl.Execute(httptest.NewRecorder(), evt.Data); err != nil {
			t.Errorf("notification template %s render error: %v", *name, err)
		}
	}
}

// TestingT is an interface wrapper for testing.T to allow helpers.
type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
}
