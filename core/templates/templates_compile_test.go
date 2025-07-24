package templates_test

import (
	"embed"
	"github.com/arran4/goa4web/core/common"
	"html/template"
	"io/fs"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed site/*.gohtml site/*/*.gohtml email/*.gohtml
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	template.Must(template.New("").Funcs(cd.Funcs(r)).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml", "email/*.gohtml"))
}

func TestParseEachTemplate(t *testing.T) {
	err := fs.WalkDir(testTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			cd := &common.CoreData{}
			if _, err := template.New("").Funcs(cd.Funcs(r)).ParseFS(testTemplates, path); err != nil {
				t.Errorf("failed to parse %s: %v", path, err)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
