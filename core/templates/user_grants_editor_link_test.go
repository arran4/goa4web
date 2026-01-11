package templates_test

import (
	"bytes"
	"database/sql"
	"embed"
	"html/template"
	"strings"
	"testing"

	admin "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/internal/db"
)

//go:embed site/admin/userGrantsEditor.gohtml
var userGrantsEditorTemplate embed.FS

func TestUserGrantsEditor_ItemIDLink(t *testing.T) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return "" },
		"assetHash": func(s string) string { return s },
	}).ParseFS(userGrantsEditorTemplate, "site/admin/userGrantsEditor.gohtml"))

	data := struct {
		User        *db.SystemGetUserByIDRow
		GrantGroups []admin.GrantGroup
	}{
		User: &db.SystemGetUserByIDRow{Idusers: 1, Username: sql.NullString{String: "u", Valid: true}},
		GrantGroups: []admin.GrantGroup{
			{
				Section: "forum",
				Item:    "topic",
				ItemID:  sql.NullInt32{Int32: 42, Valid: true},
				Link:    "/admin/forum/topics/topic/42",
			},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "userGrantsEditor.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a href="/admin/forum/topics/topic/42">42</a>`) {
		t.Fatalf("expected link in output, got %s", html)
	}
}
