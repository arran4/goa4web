package notifications

import (
	"bytes"
	htemplate "html/template"
	"os"
	"strings"
	"testing"
	ttemplate "text/template"
)

// parseTxtar parses a txtar archive into a map of filenames to content.
func parseTxtar(data []byte) map[string][]byte {
	files := make(map[string][]byte)
	lines := bytes.Split(data, []byte("\n"))
	var currentFile string
	var currentContent []byte
	inContent := false

	for _, line := range lines {
		// Check for marker
		lineTrimmed := bytes.TrimRight(line, "\r")
		if bytes.HasPrefix(lineTrimmed, []byte("-- ")) && bytes.HasSuffix(lineTrimmed, []byte(" --")) {
			if inContent {
				// Store previous file
				files[currentFile] = currentContent
			}
			currentFile = string(lineTrimmed[3 : len(lineTrimmed)-3])
			currentContent = nil
			inContent = true
			continue
		}
		if inContent {
			if len(currentContent) > 0 {
				currentContent = append(currentContent, '\n')
			}
			currentContent = append(currentContent, line...)
		}
	}
	if inContent {
		files[currentFile] = currentContent
	}
	return files
}

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

	// Read email txtar once
	emailTmplPath := "../../core/templates/email/adminNotificationForumThreadCreateEmail.txtar"
	emailTxtarData, err := os.ReadFile(emailTmplPath)
	if err != nil {
		t.Fatalf("failed to read email template txtar: %v", err)
	}
	emailFiles := parseTxtar(emailTxtarData)

	// 2. Email Text Template (.gotxt)
	t.Run("Email Text Template", func(t *testing.T) {
		content, ok := emailFiles["body.gotxt"]
		if !ok {
			t.Fatalf("body.gotxt not found in txtar")
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
		content, ok := emailFiles["body.gohtml"]
		if !ok {
			t.Fatalf("body.gohtml not found in txtar")
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
