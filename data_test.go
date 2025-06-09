package main

import (
	"embed"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

//go:embed templates/*.gohtml
var testTemplates embed.FS

func TestCompileGoHTML(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
  template.Must(template.New("").Funcs(NewFuncs(r)).ParseFS(os.DirFS("./templates"), "*.gohtml"))
}
