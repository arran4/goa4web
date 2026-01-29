package templates_test

import (
	"bytes"
	"database/sql"
	"embed"
	"html/template"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

//go:embed site/admin/grantPage.gohtml site/admin/grantsPage.gohtml
var grantTemplates embed.FS

type grantWithNames struct {
	*db.Grant
	UserName string
	RoleName string
	ItemLink string
}

type grantAction struct {
	ID     int32
	Name   string
	Active bool
}

type grantGroup struct {
	*db.Grant
	UserName string
	RoleName string
	ItemLink string
	Actions  []grantAction
}

func TestGrantPageLinks(t *testing.T) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return "" },
	}).ParseFS(grantTemplates, "site/admin/grantPage.gohtml"))
	template.Must(tmpl.New("head").Parse(""))
	template.Must(tmpl.New("tail").Parse(""))

	g := &db.Grant{
		ID:      1,
		UserID:  sql.NullInt32{Int32: 5, Valid: true},
		RoleID:  sql.NullInt32{Int32: 7, Valid: true},
		Section: "forum",
		Item:    sql.NullString{String: "topic", Valid: true},
		ItemID:  sql.NullInt32{Int32: 42, Valid: true},
	}
	data := struct{ Grant grantWithNames }{
		Grant: grantWithNames{
			Grant:    g,
			UserName: "bob",
			RoleName: "admin",
			ItemLink: "/admin/forum/topics/topic/42",
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "grantPage.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a href="/admin/user/5">bob (5)</a>`) {
		t.Fatalf("expected user link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/role/7">admin (7)</a>`) {
		t.Fatalf("expected role link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/forum/topics/topic/42">topic</a>`) {
		t.Fatalf("expected item link, got %s", html)
	}
}

type mockFilter struct {
	Username string
	RoleName string
	Section  string
	Item     string
	ItemID   string
	Active   string
	Sort     string
	Dir      string
}

func TestGrantsPageLinks(t *testing.T) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return "" },
	}).ParseFS(grantTemplates, "site/admin/grantsPage.gohtml"))
	template.Must(tmpl.New("head").Parse(""))
	template.Must(tmpl.New("tail").Parse(""))

	g := &db.Grant{
		ID:      1,
		UserID:  sql.NullInt32{Int32: 5, Valid: true},
		RoleID:  sql.NullInt32{Int32: 7, Valid: true},
		Section: "forum",
		Item:    sql.NullString{String: "topic", Valid: true},
		ItemID:  sql.NullInt32{Int32: 42, Valid: true},
		Active:  true,
	}
	data := struct {
		Grants []grantGroup
		Filter mockFilter
	}{
		Grants: []grantGroup{
			{
				Grant:    g,
				UserName: "bob",
				RoleName: "admin",
				ItemLink: "/admin/forum/topics/topic/42",
				Actions:  []grantAction{{ID: 1, Name: "search", Active: true}},
			},
		},
		Filter: mockFilter{},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "grantsPage.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a href="/admin/user/5">bob (5)</a>`) {
		t.Fatalf("expected user link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/role/7">admin (7)</a>`) {
		t.Fatalf("expected role link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/forum/topics/topic/42">topic</a>`) {
		t.Fatalf("expected item link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/grant/1" class="pill">search</a>`) {
		t.Fatalf("expected action pill, got %s", html)
	}
}

func TestGrantsPageLinksAnyone(t *testing.T) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return "" },
	}).ParseFS(grantTemplates, "site/admin/grantsPage.gohtml"))
	template.Must(tmpl.New("head").Parse(""))
	template.Must(tmpl.New("tail").Parse(""))

	g := &db.Grant{ID: 1, Active: true}
	data := struct {
		Grants []grantGroup
		Filter mockFilter
	}{
		Grants: []grantGroup{
			{
				Grant:    g,
				UserName: "Anyone",
				Actions:  []grantAction{{ID: 1, Name: "search", Active: true}},
			},
		},
		Filter: mockFilter{},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "grantsPage.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a href="/admin/grants/anyone">Anyone</a>`) {
		t.Fatalf("expected anyone link, got %s", html)
	}
	if !strings.Contains(html, `<a href="/admin/grant/1" class="pill">search</a>`) {
		t.Fatalf("expected action pill, got %s", html)
	}
}
