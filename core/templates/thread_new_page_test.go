package templates

import (
	"bytes"
	"errors"
	"html/template"
	"strings"
	"testing"
	"time"
)

// Define types for template context
type MockUser struct {
	Username struct{ String string }
}
type MockComment struct {
	Written struct{ Time time.Time }
}
type MockCoreData struct {
	CurrentUserLoaded             *MockUser
	SelectedSectionThreadComments []*MockComment
	Offset                        int
	Location                      string
}

func (m *MockCoreData) LocalTime(t time.Time) time.Time {
	return t
}

func TestThreadNewPageJS(t *testing.T) {
	funcMap := template.FuncMap{
		"csrfField": func() template.HTML { return "" },
		"assetHash": func(s string) string { return s },
		"cd": func() *MockCoreData {
			return &MockCoreData{
				Location: "UTC",
			}
		},
		"add": func(a, b int) int { return a + b },
		"now": func() time.Time { return time.Now() },
		"since": func(t1, t2 time.Time) string { return "" },
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			m := make(map[string]interface{}, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				k, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				m[k] = values[i+1]
			}
			return m, nil
		},
	}

	tmpl := template.New("test").Funcs(funcMap)

	// Stub templates
	stubs := []string{
		`{{define "head"}}{{end}}`,
		`{{define "tail"}}{{end}}`,
		`{{define "a4codeControls"}}{{end}}`,
		`{{define "languageCombobox"}}{{end}}`,
	}
	for _, stub := range stubs {
		if _, err := tmpl.Parse(stub); err != nil {
			t.Fatalf("parse stub: %v", err)
		}
	}

	if _, err := tmpl.ParseFiles("site/forum/threadNewPage.gohtml"); err != nil {
		t.Fatalf("parse templates: %v", err)
	}

	tests := []struct {
		name     string
		basePath string
		wantSrc  string
	}{
		{
			name:     "Private Forum",
			basePath: "/private",
			wantSrc:  `<script src="/private/topic_labels.js"></script>`,
		},
		{
			name:     "Public Forum",
			basePath: "/forum",
			wantSrc:  `<script src="/forum/topic_labels.js"></script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := struct {
				BasePath string
				Topic    struct{ Idforumtopic int32 }
				QuoteText string
			}{
				BasePath: tt.basePath,
			}
			data.Topic.Idforumtopic = 16

			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, "threadNewPage.gohtml", data); err != nil {
				t.Fatalf("execute template: %v", err)
			}

			out := buf.String()
			if !strings.Contains(out, tt.wantSrc) {
				t.Errorf("Expected script tag %q not found in output. Output:\n%s", tt.wantSrc, out)
			} else {
				t.Logf("Confirmed: script tag is %q", tt.wantSrc)
			}
		})
	}
}
