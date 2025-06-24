//go:build live
// +build live

package templates

import (
	"html/template"
	"os"
)

func GetCompiledTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(
		template.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/templates"),
			"*.gohtml", "*/*.gohtml"))
}

func GetMainCSSData() []byte {
	b, err := os.ReadFile("core/templates/assets/main.css")
	if err != nil {
		panic(err)
	}
	return b
}
