//go:build !live
// +build !live

package templates

import (
	"embed"
	htemplate "html/template"
	ttemplate "text/template"
)

var (
	//go:embed "site/*.gohtml" "site/*/*.gohtml"
	siteTemplatesFS embed.FS
	//go:embed "notifications/*.gotxt"
	notificationTemplatesFS embed.FS
	//go:embed "email/*.gohtml"
	emailHtmlTemplatesFS embed.FS
	//go:embed "email/*.gotxt"
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

func GetCompiledSiteTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(siteTemplatesFS, "site/*.gohtml", "site/*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(notificationTemplatesFS, "notifications/*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(emailHtmlTemplatesFS, "email/*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(emailTextTemplatesFS, "email/*.gotxt"))
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
