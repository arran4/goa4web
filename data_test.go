package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCompileGoHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	template.Must(template.New("").Funcs(NewFuncs(req)).ParseFS(os.DirFS("./templates"), "*.gohtml"))
}
