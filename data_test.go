package main

import (
	"html/template"
	"os"
	"testing"
)

func TestCompileGoHTML(t *testing.T) {
	template.Must(template.New("").Funcs(NewFuncs(r)).ParseFS(os.DirFS("./templates"), "*.gohtml"))
}
