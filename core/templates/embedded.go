//go:build !live
// +build !live

package templates

import (
	"embed"
	"html/template"
)

var (
	//go:embed "site/*.gohtml" "site/*/*.gohtml"
	siteTemplatesFS embed.FS
	//go:embed "notifications/*.txt"
	notificationTemplatesFS embed.FS
	//go:embed "email/*.html"
	emailHtmlTemplatesFS embed.FS
	//go:embed "email/*.txt"
	emailTextTemplatesFS embed.FS
	//go:embed "assets/main.css"
	mainCSSData []byte
	//go:embed "assets/favicon.svg"
	faviconData []byte
	//go:embed "assets/pasteimg.js"
	pasteImageJSData []byte
	//go:embed "assets/notifications.js"
	notificationsJSData []byte
)

func GetCompiledSiteTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(siteTemplatesFS, "site/*.gohtml", "site/*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(notificationTemplatesFS, "notifications/*.txt"))
}

func GetCompiledEmailHtmlTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(emailHtmlTemplatesFS, "email/*.html"))
}

func GetCompiledEmailTextTemplates(funcs template.FuncMap) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(emailTextTemplatesFS, "email/*.txt"))
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
