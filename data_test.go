package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

//go:embed templates/*.gohtml
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	template.Must(template.New("").Funcs(NewFuncs(r)).ParseFS(testTemplates, "templates/*.gohtml"))
}

func TestParseEachTemplate(t *testing.T) {
	entries, err := fs.ReadDir(testTemplates, "templates")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gohtml") {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			path := "templates/" + e.Name()
			if _, err := template.New("").Funcs(NewFuncs(r)).ParseFS(testTemplates, path); err != nil {
				t.Errorf("failed to parse %s: %v", e.Name(), err)
			}
		})
	}
}
