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

//go:embed site/admin/roleGrantsEditor.gohtml
var roleGrantsEditorTemplate embed.FS

func TestRoleGrantsEditor_ItemIDLink(t *testing.T) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return "" },
		"assetHash": func(s string) string { return s },
	}).ParseFS(roleGrantsEditorTemplate, "site/admin/roleGrantsEditor.gohtml"))

	data := struct {
		Role        *db.Role
		GrantGroups []admin.GrantGroup
	}{
		Role: &db.Role{ID: 1, Name: "test"},
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
	if err := tmpl.ExecuteTemplate(&buf, "roleGrantsEditor.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a href="/admin/forum/topics/topic/42">42</a>`) {
		t.Fatalf("expected link in output, got %s", html)
	}
}
