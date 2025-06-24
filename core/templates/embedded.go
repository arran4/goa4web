//go:build !live
// +build !live

package templates

import (
	"embed"
	"html/template"
)

var (
	//go:embed "templates/*.gohtml" "templates/*/*.gohtml"
	templateFS embed.FS
	//go:embed "assets/main.css"
	mainCSSData []byte
)

func GetCompiledTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/*.gohtml", "templates/*/*.gohtml"))
}

func GetMainCSSData() []byte {
	return mainCSSData
}
