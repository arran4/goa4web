package notifications

import (
	"bytes"
	htemplate "html/template"
	"os"
	"strings"
	"testing"
	ttemplate "text/template"
)

func TestAdminNotificationTemplate(t *testing.T) {
	// 1. Notification Template (.gotxt)
	t.Run("Notification Template", func(t *testing.T) {
		tmplPath := "../../core/templates/notifications/adminNotificationForumThreadCreateEmail.gotxt"
		content, err := os.ReadFile(tmplPath)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		tmpl, err := ttemplate.New("notif").Parse(string(content))
		if err != nil {
			t.Fatalf("failed to parse template: %v", err)
		}

		data := map[string]any{
			"Item": map[string]any{
				"Username":            "arran",
				"SubjectPrefix":       "Private Forum",
				"TopicTitle":          "My Topic",
				"ThreadOpenerPreview": "Hello world",
			},
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}

		expected := "User arran created a new private forum thread in My Topic: Hello world"
		if strings.TrimSpace(buf.String()) != expected {
			t.Errorf("Private: expected %q, got %q", expected, buf.String())
		}

		// Test public forum
		data["Item"].(map[string]any)["SubjectPrefix"] = "Forum"
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}
		expectedPublic := "User arran created a new forum thread in My Topic: Hello world"
		if strings.TrimSpace(buf.String()) != expectedPublic {
			t.Errorf("Public: expected %q, got %q", expectedPublic, buf.String())
		}
	})

	// 2. Email Text Template (.gotxt)
	t.Run("Email Text Template", func(t *testing.T) {
		tmplPath := "../../core/templates/email/adminNotificationForumThreadCreateEmail.gotxt"
		content, err := os.ReadFile(tmplPath)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		tmpl, err := ttemplate.New("email_text").Parse(string(content))
		if err != nil {
			t.Fatalf("failed to parse template: %v", err)
		}

		data := map[string]any{
			"Item": map[string]any{
				"Username":            "arran",
				"SubjectPrefix":       "Private Forum",
				"TopicTitle":          "My Topic",
				"ThreadOpenerPreview": "Hello world",
				"ThreadURL":           "http://thread",
			},
			"UnsubscribeUrl": "http://unsub",
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}

		expectedContent := "User arran created a new private forum thread in My Topic: Hello world"
		if !strings.Contains(buf.String(), expectedContent) {
			t.Errorf("Private: expected to contain %q, got %q", expectedContent, buf.String())
		}

		// Test public forum
		data["Item"].(map[string]any)["SubjectPrefix"] = "Forum"
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}
		expectedPublic := "User arran created a new forum thread in My Topic: Hello world"
		if !strings.Contains(buf.String(), expectedPublic) {
			t.Errorf("Public: expected to contain %q, got %q", expectedPublic, buf.String())
		}
	})

	// 3. Email HTML Template (.gohtml)
	t.Run("Email HTML Template", func(t *testing.T) {
		tmplPath := "../../core/templates/email/adminNotificationForumThreadCreateEmail.gohtml"
		content, err := os.ReadFile(tmplPath)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		tmpl, err := htemplate.New("email_html").Parse(string(content))
		if err != nil {
			t.Fatalf("failed to parse template: %v", err)
		}

		data := map[string]any{
			"Item": map[string]any{
				"Username":            "arran",
				"SubjectPrefix":       "Private Forum",
				"TopicTitle":          "My Topic",
				"ThreadOpenerPreview": "Hello world",
				"ThreadURL":           "http://thread",
			},
			"UnsubscribeUrl": "http://unsub",
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}

		expectedContent := "User arran created a new private forum thread in My Topic: Hello world"
		if !strings.Contains(buf.String(), expectedContent) {
			t.Errorf("Private: expected to contain %q, got %q", expectedContent, buf.String())
		}

		// Test public forum
		data["Item"].(map[string]any)["SubjectPrefix"] = "Forum"
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}
		expectedPublic := "User arran created a new forum thread in My Topic: Hello world"
		if !strings.Contains(buf.String(), expectedPublic) {
			t.Errorf("Public: expected to contain %q, got %q", expectedPublic, buf.String())
		}
	})
}
