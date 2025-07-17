//go:build live
// +build live

package templates

import (
	"html/template"
	"os"
)

func GetCompiledSiteTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/site"), "*.gohtml", "*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/notifications"), "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/email"), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/email"), "*.gotxt"))
}

func GetMainCSSData() []byte {
	b, err := os.ReadFile("core/templates/assets/main.css")
	if err != nil {
		panic(err)
	}
	return b
}

func GetFaviconData() []byte {
	b, err := os.ReadFile("core/templates/assets/favicon.svg")
	if err != nil {
		panic(err)
	}
	return b
}

func GetPasteImageJSData() []byte {
	b, err := os.ReadFile("core/templates/assets/pasteimg.js")
	if err != nil {
		panic(err)
	}
	return b
}

func GetNotificationsJSData() []byte {
	b, err := os.ReadFile("core/templates/assets/notifications.js")
	if err != nil {
		panic(err)
	}
	return b
}
