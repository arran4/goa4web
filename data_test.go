package main

import (
	"html/template"
	"os"
	"testing"
)

func TestCompileGoHTML(t *testing.T) {
	template.Must(template.New("").Funcs(NewFuncs()).ParseFS(os.DirFS("./templates"), "*.gohtml"))
}
