//go:build live
// +build live

package main

import (
	"html/template"
	"os"
)

func getCompiledTemplates() *template.Template {
	return template.Must(template.New("").Funcs(NewFuncs()).ParseFS(os.DirFS("./templates"), "*.tmpl"))
}

func getMainCSSData() []byte {
	b, err := os.ReadFile("main.css")
	if err != nil {
		panic(err)
	}
	return b
}
