package templates_test

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestHeadTemplateRendersSiteTitle(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{SiteTitle: "My Site", PageTitle: "Page"}
	tmpl := template.Must(template.New("").Funcs(cd.Funcs(r)).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml"))
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "head", nil); err != nil {
		t.Fatalf("execute head: %v", err)
	}
	if !strings.Contains(b.String(), "<title>Page - My Site</title>") {
		t.Fatalf("unexpected output: %s", b.String())
	}
}
