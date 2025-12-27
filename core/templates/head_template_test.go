package templates_test

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestHeadTemplateRendersSiteTitle(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{SiteTitle: "My Site", PageTitle: "Page"}
	funcs := cd.Funcs(r)
	funcs["assetHash"] = func(s string) string { return s }
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml"))
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "head", nil); err != nil {
		t.Fatalf("execute head: %v", err)
	}
	if !strings.Contains(b.String(), "<title>Page - My Site</title>") {
		t.Fatalf("unexpected output: %s", b.String())
	}
}

func TestHeadTemplateIncludesModuleScripts(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := common.NewCoreData(r.Context(), nil, config.NewRuntimeConfig(),
		common.WithSiteTitle("My Site"),
		common.WithRouterModules([]string{"images", "websocket"}),
	)
	cd.UserID = 1
	cd.PageTitle = "Page"
	funcs := cd.Funcs(r)
	funcs["assetHash"] = func(s string) string { return s }
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml"))
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "head", nil); err != nil {
		t.Fatalf("execute head: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `<script src="/images/pasteimg.js"></script>`) {
		t.Errorf("missing images script: %s", out)
	}
	if !strings.Contains(out, `<script src="/websocket/notifications.js"></script>`) {
		t.Errorf("missing notifications script: %s", out)
	}
}
