package faq

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func TestHappyPathAdminTemplatesPageRender(t *testing.T) {
	// 1. Setup Data
	languages := []*db.Language{
		{ID: 1, Nameof: sql.NullString{String: "English", Valid: true}},
	}
	data := AdminTemplatesPageData{
		Languages:        languages,
		Templates:        []string{"test-template"},
		SelectedTemplate: "test-template",
	}

	// 2. Setup Funcs
	cd := &common.CoreData{}
	r := httptest.NewRequest("GET", "/", nil)
	funcs := cd.Funcs(r)

	// Load templates from disk (relative to this test file)
	tmpl := templates.GetCompiledSiteTemplates(funcs, templates.WithDir("../../core/templates"))

	// 3. Execute
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "faq/adminTemplatesPage.gohtml", data)

	// 4. Verification
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if !strings.Contains(buf.String(), "English") {
		t.Errorf("Expected output to contain 'English', but got: %s", buf.String())
	}
}
