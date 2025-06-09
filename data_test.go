package main

import (
	"embed"
	"html/template"
	"net/http/httptest"
	"testing"
)

//go:embed templates/*.gohtml
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	template.Must(template.New("").Funcs(NewFuncs(r)).ParseFS(testTemplates, "templates/*.gohtml"))
}
