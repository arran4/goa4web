package templates_test

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestPaginationTemplateWithoutPageSize(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{Config: config.NewRuntimeConfig()}
	cd.PrevLink = "/prev"
	funcs := cd.Funcs(r)
	funcs["localTime"] = func(t time.Time) time.Time { return t }
	funcs["assetHash"] = func(s string) string { return s }
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml", "email/*.gohtml"))
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "tail", cd); err != nil {
		t.Fatalf("execute tail template: %v", err)
	}
}
