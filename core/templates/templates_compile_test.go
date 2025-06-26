package templates

import (
	"embed"
	corecommon "github.com/arran4/goa4web/core/common"
	"html/template"
	"io/fs"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed templates/*.gohtml templates/*/*.gohtml
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	template.Must(template.New("").Funcs(corecommon.NewFuncs(r)).ParseFS(testTemplates,
		"templates/*.gohtml", "templates/*/*.gohtml"))
}

func TestParseEachTemplate(t *testing.T) {
	err := fs.WalkDir(testTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if _, err := template.New("").Funcs(corecommon.NewFuncs(r)).ParseFS(testTemplates, path); err != nil {
				t.Errorf("failed to parse %s: %v", path, err)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
