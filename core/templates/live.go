//go:build live
// +build live

package templates

import (
	htemplate "html/template"
	"log"
	"os"
	ttemplate "text/template"
)

func init() {
	log.Printf("Live Template Mode")
}

func GetCompiledSiteTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/site"), "*.gohtml", "*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/notifications"), "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/email"), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/email"), "*.gotxt"))
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

// GetRoleGrantsEditorJSData returns the JavaScript powering the
// role grants drag-and-drop editor.
func GetRoleGrantsEditorJSData() []byte {
	b, err := os.ReadFile("core/templates/assets/role_grants_editor.js")
	if err != nil {
		panic(err)
	}
	return b
}
