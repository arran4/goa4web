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
	//go:embed "assets/favicon.svg"
	faviconData []byte
	//go:embed "assets/pasteimg.js"
	pasteImageJSData []byte
	//go:embed "assets/notifications.js"
	notificationsJSData []byte
)

func GetCompiledTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/*.gohtml", "templates/*/*.gohtml"))
}

func GetMainCSSData() []byte {
	return mainCSSData
}

// GetFaviconData returns the site's favicon image data.
func GetFaviconData() []byte {
	return faviconData
}

// GetPasteImageJSData returns the JavaScript that enables image pasting.
func GetPasteImageJSData() []byte {
	return pasteImageJSData
}

// GetNotificationsJSData returns the JavaScript used for real-time
// notification updates.
func GetNotificationsJSData() []byte { return notificationsJSData }
