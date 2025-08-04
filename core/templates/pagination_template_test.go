package templates_test

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestPaginationTemplateWithoutPageSize(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	cd.PrevLink = "/prev"
	tmpl := template.Must(template.New("").Funcs(cd.Funcs(r)).ParseFS(testTemplates,
		"site/*.gohtml", "site/*/*.gohtml", "email/*.gohtml"))
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "tail", struct{}{}); err != nil {
		t.Fatalf("execute tail template: %v", err)
	}
}
