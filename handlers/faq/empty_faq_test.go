package faq_test

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

func stubFuncs() template.FuncMap {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{Config: config.NewRuntimeConfig()}
	return cd.Funcs(req)
}

func TestEmptyFAQMessage(t *testing.T) {
	// Load templates from disk so we can modify them and verify the fix
	tmpl := templates.GetCompiledSiteTemplates(stubFuncs(), templates.WithDir("../../core/templates"))

	cd := &common.CoreData{
		Config: config.NewRuntimeConfig(),
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "faq/page.gohtml", cd); err != nil {
		t.Fatalf("render faq/page.gohtml: %v", err)
	}

	output := buf.String()

	expectedMessage := "There are no questions at the moment."

	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected message '%s' not found in output", expectedMessage)
	}
}
