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
		"site/*/*.gohtml", "site/*/*/*.gohtml"))
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
		"site/*/*.gohtml", "site/*/*/*.gohtml"))
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

func TestHeadTemplateIncludesBFCacheSafeguardOnlyForSensitivePages(t *testing.T) {
	tests := []struct {
		name      string
		userID    int32
		pageTitle string
		want      bool
	}{
		{name: "authenticated page", userID: 1, pageTitle: "Account", want: true},
		{name: "anonymous login", pageTitle: "Login", want: true},
		{name: "anonymous public page", pageTitle: "News", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(r.Context(), nil, config.NewRuntimeConfig(), common.WithSiteTitle("My Site"))
			cd.UserID = tt.userID
			cd.PageTitle = tt.pageTitle
			funcs := cd.Funcs(r)
			funcs["assetHash"] = func(s string) string { return s }
			tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(testTemplates,
				"site/*/*.gohtml", "site/*/*/*.gohtml"))

			var b strings.Builder
			if err := tmpl.ExecuteTemplate(&b, "head", nil); err != nil {
				t.Fatalf("execute head: %v", err)
			}
			out := b.String()
			hasListener := strings.Contains(out, "window.addEventListener('pageshow'")
			if hasListener != tt.want {
				t.Errorf("pageshow listener present = %t; want %t; output: %s", hasListener, tt.want, out)
			}
			if tt.want {
				normalized := strings.Join(strings.Fields(out), " ")
				persistedReload := "window.addEventListener('pageshow', function(event) { if (event.persisted) { window.location.reload(); } });"
				if !strings.Contains(normalized, persistedReload) || strings.Count(out, "window.location.reload()") != 1 {
					t.Errorf("BFCache safeguard must reload only from the persisted restoration branch: %s", out)
				}
			}
		})
	}
}
