//go:build !live
// +build !live

package main

import (
	"embed"
	"html/template"
)

var (
	//go:embed "templates/*.gohtml"
	templateFS        embed.FS
	compiledTemplates = template.Must(template.New("").Funcs(NewFuncs()).ParseFS(templateFS, "templates/*.gohtml"))
	//go:embed "main.css"
	mainCSSData []byte
)

func getCompiledTemplates() *template.Template {
	return compiledTemplates
}

func getMainCSSData() []byte {
	return mainCSSData
}
