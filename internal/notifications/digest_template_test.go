package notifications

import (
	"database/sql"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

func TestDigestTemplateExecution(t *testing.T) {
	// Mimic the data structure in digest_worker.go
	notifs := []*db.Notification{
		{
			Message:   sql.NullString{String: "Test Message", Valid: true},
			CreatedAt: time.Now(),
			Link:      sql.NullString{String: "/test", Valid: true},
		},
	}
	itemData := struct {
		Notifications []*db.Notification
		BaseURL       string
		LastSentAt    sql.NullTime
	}{
		Notifications: notifs,
		BaseURL:       "http://localhost",
		LastSentAt:    sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true},
	}

	data := EmailData{
		any:  itemData,
		Item: itemData,
	}

	// Updated template content (from body.gotxt)
	tmplBody := `Here is your daily digest of unread notifications:

{{ range .Item.Notifications }}
* {{ .Message.String }} ({{ .CreatedAt.Format "2006-01-02 15:04" }})
  {{ if .Link.Valid }}{{ $.Item.BaseURL }}{{ .Link.String }}{{ end }}
{{ end }}

You have {{ len .Item.Notifications }} unread notifications.
Visit {{ .Item.BaseURL }}/usr/notifications to view them all.

{{ if .Item.LastSentAt.Valid }}
Last digest sent at: {{ .Item.LastSentAt.Time.Format "2006-01-02 15:04" }}
{{ end }}
`

	tmpl, err := template.New("digest").Parse(tmplBody)
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Message") {
		t.Errorf("Expected 'Test Message', got %q", output)
	}
	if !strings.Contains(output, "http://localhost/test") {
		t.Errorf("Expected 'http://localhost/test', got %q", output)
	}
	if !strings.Contains(output, "Last digest sent at:") {
		t.Errorf("Expected 'Last digest sent at:', got %q", output)
	}
}
