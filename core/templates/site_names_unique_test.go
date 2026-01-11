package templates_test

import (
	"embed"
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

//go:embed site/*.gohtml site/*/*.gohtml
var uniqueTemplates embed.FS

func TestSiteTemplateNamesUnique(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	funcs["assetHash"] = func(s string) string { return s }
	tmpl, err := template.New("").Funcs(funcs).ParseFS(uniqueTemplates,
		"site/*.gohtml", "site/*/*.gohtml")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	seen := map[string]string{}
	for _, tt := range tmpl.Templates() {
		name := tt.Name()
		if name == "" {
			continue
		}
		parseName := ""
		if tt.Tree != nil {
			parseName = tt.Tree.ParseName
		}
		if prev, ok := seen[name]; ok {
			t.Fatalf("duplicate template name %q found in %s and %s", name, prev, parseName)
		}
		seen[name] = parseName
	}
}
