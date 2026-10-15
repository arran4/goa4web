package templates_test

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

//go:embed all:site
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	funcs["localTime"] = func(t time.Time) time.Time { return t }
	funcs["assetHash"] = func(s string) string { return s }
	root := template.New("").Funcs(funcs)
	err := fs.WalkDir(testTemplates, "site", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(name, ".gohtml") {
			return nil
		}
		body, err := fs.ReadFile(testTemplates, name)
		if err != nil {
			return err
		}
		_, err = root.New(strings.TrimPrefix(name, "site/")).Parse(string(body))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestCompiledSiteTemplatesContainEveryEmbeddedTemplate guards recursive site
// template discovery as the directory layout evolves.
func TestCompiledSiteTemplatesContainEveryEmbeddedTemplate(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	compiled := templates.GetCompiledSiteTemplates(cd.Funcs(r), templates.WithSilence(true))

	for _, name := range templates.ListSiteTemplateNames(templates.WithSilence(true)) {
		if compiled.Lookup(name) == nil {
			t.Errorf("compiled templates are missing %q", name)
		}
	}
}

func TestCompiledSiteTemplatesRenderIncludeViaInjectedRenderer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	var compiled *template.Template
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithSiteTemplateRenderer(func(_ *common.CoreData, w io.Writer, _ *http.Request, name string, data any) error {
		return compiled.ExecuteTemplate(w, name, data)
	}))
	compiled = templates.GetCompiledSiteTemplates(cd.Funcs(req), templates.WithSilence(true))
	cd.CustomIndexItems = []common.IndexItem{{TemplateName: "footer"}}

	var rendered bytes.Buffer
	if err := compiled.ExecuteTemplate(&rendered, "index", cd); err != nil {
		t.Fatalf("render compiled index template: %v", err)
	}
	if !strings.Contains(rendered.String(), "Date and time is now:") {
		t.Fatalf("included footer was not rendered: %s", rendered.String())
	}
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
			funcs := cd.Funcs(r)
			funcs["localTime"] = func(t time.Time) time.Time { return t }
			funcs["assetHash"] = func(s string) string { return s }
			if _, err := template.New("").Funcs(funcs).ParseFS(testTemplates, path); err != nil {
				t.Errorf("failed to parse %s: %v", path, err)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
