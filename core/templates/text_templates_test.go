package templates_test

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

//go:embed notifications/*.gotxt email/*.gotxt
var textTemplates embed.FS

func TestParseGoTxtTemplates(t *testing.T) {
	emailTemplates := templates.GetCompiledEmailTextTemplates(nil)
	notificationTemplates := templates.GetCompiledNotificationTemplates(nil)

	err := fs.WalkDir(textTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".gotxt") {
			return nil
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			switch {
			case strings.HasPrefix(path, "email/"):
				name := strings.TrimPrefix(path, "email/")
				if emailTemplates.Lookup(name) == nil {
					t.Errorf("failed to parse %s", path)
				}
			case strings.HasPrefix(path, "notifications/"):
				name := strings.TrimPrefix(path, "notifications/")
				if notificationTemplates.Lookup(name) == nil {
					t.Errorf("failed to parse %s", path)
				}
			default:
				t.Fatalf("unexpected template path %s", path)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnouncementTemplatesExist(t *testing.T) {
	nt := templates.GetCompiledNotificationTemplates(nil)
	if nt.Lookup("announcement.gotxt") == nil {
		t.Fatalf("missing announcement notification template")
	}
	et := templates.GetCompiledEmailHtmlTemplates(nil)
	if et.Lookup("announcementEmail.gohtml") == nil {
		t.Fatalf("missing announcement email html template")
	}
	tt := templates.GetCompiledEmailTextTemplates(nil)
	if tt.Lookup("announcementEmail.gotxt") == nil {
		t.Fatalf("missing announcement email text template")
	}
}

func TestAllEmailTemplatesComplete(t *testing.T) {
	htmlT := templates.GetCompiledEmailHtmlTemplates(nil)
	textT := templates.GetCompiledEmailTextTemplates(nil)

	type trio struct{ html, text, subj bool }
	m := map[string]*trio{}

	for _, tmpl := range htmlT.Templates() {
		name := tmpl.Name()
		if !strings.HasSuffix(name, ".gohtml") {
			continue
		}
		prefix := strings.TrimSuffix(name, ".gohtml")
		tr := m[prefix]
		if tr == nil {
			tr = &trio{}
			m[prefix] = tr
		}
		tr.html = true
	}

	for _, tmpl := range textT.Templates() {
		name := tmpl.Name()
		switch {
		case strings.HasSuffix(name, "Subject.gotxt"):
			prefix := strings.TrimSuffix(name, "Subject.gotxt")
			tr := m[prefix]
			if tr == nil {
				tr = &trio{}
				m[prefix] = tr
			}
			tr.subj = true
		case strings.HasSuffix(name, ".gotxt"):
			prefix := strings.TrimSuffix(name, ".gotxt")
			tr := m[prefix]
			if tr == nil {
				tr = &trio{}
				m[prefix] = tr
			}
			tr.text = true
		}
	}

	for p, tr := range m {
		if !(tr.html && tr.text && tr.subj) {
			t.Errorf("template set %s incomplete: html=%v text=%v subj=%v", p, tr.html, tr.text, tr.subj)
		}
	}
}
