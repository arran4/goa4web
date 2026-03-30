package templates_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"net/http/httptest"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

func TestTemplateRegression(t *testing.T) {
	config := &config.RuntimeConfig{}

	cd := &common.CoreData{
		Config: config,
		AdminMode: true,
		CustomIndexItems: []common.IndexItem{
			{Name: "Item 1", Link: "/item1", Folded: false},
			{Name: "Item 2", Link: "/item2", Folded: true},
		},
	}

	testCases := []struct {
		name         string
		templateName string
		data         interface{}
	}{
		{
			name:         "index",
			templateName: "index",
			data:         map[string]interface{}{"cd": cd},
		},
		{
			name:         "loginPage",
			templateName: "pages/auth/loginPage.gohtml",
			data:         map[string]interface{}{"cd": cd, "redirect": "/admin"},
		},
		{
			name:         "header",
			templateName: "header",
			data:         map[string]interface{}{"cd": cd},
		},
		{
			name:         "footer",
			templateName: "footer",
			data:         map[string]interface{}{"cd": cd},
		},
		{
			name:         "head",
			templateName: "head",
			data:         map[string]interface{}{"cd": cd},
		},
		{
			name:         "notFoundPage",
			templateName: "pages/misc/notFoundPage.gohtml",
			data:         map[string]interface{}{"cd": cd},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			funcs := cd.Funcs(r)
			funcs["localTime"] = func(t time.Time) time.Time { return t }
			funcs["assetHash"] = func(s string) string { return s }
			tmpl := templates.GetCompiledSiteTemplates(funcs)
			if tmpl == nil {
				t.Fatalf("Failed to get compiled site templates")
			}

			var buf bytes.Buffer
			err := tmpl.ExecuteTemplate(&buf, tc.templateName, tc.data)
			if err != nil {
				t.Fatalf("Failed to execute template %s: %v", tc.templateName, err)
			}

			output := normalizeHTML(buf.String())

			goldenFile := filepath.Join("testdata", fmt.Sprintf("%s.golden", tc.name))
			if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
				os.MkdirAll("testdata", 0755)
				err = os.WriteFile(goldenFile, []byte(output), 0644)
				if err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Created golden file for %s", tc.name)
			} else {
				expected, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatalf("Failed to read golden file: %v", err)
				}

				if output != string(expected) {
					t.Errorf("Template %s output doesn't match golden file. \nExpected:\n%s\n\nGot:\n%s", tc.templateName, string(expected), output)
				}
			}
		})
	}
}

func normalizeHTML(input string) string {
	// Remove excessive whitespace to make comparison more robust
	lines := strings.Split(input, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}
