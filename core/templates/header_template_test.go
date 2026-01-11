package templates_test

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestHeaderTemplateRendersMobileNavIconAndText(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{
		SiteTitle: "My Site",
		PageTitle: "Page",
	}
	funcs := cd.Funcs(r)
	funcs["assetHash"] = func(s string) string { return s }
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml"))

	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "header", nil); err != nil {
		t.Fatalf("execute header: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `<span class="hamburger-icon" aria-hidden="true">&#9776;</span>`) {
		t.Fatalf("header missing hamburger icon: %s", out)
	}
	if !strings.Contains(out, `<span class="hamburger-text">Menu</span>`) {
		t.Fatalf("header missing hamburger text label: %s", out)
	}
}
