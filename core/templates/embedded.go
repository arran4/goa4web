//go:build !live
// +build !live

package templates

import (
	"embed"
	htemplate "html/template"
	"io/fs"
	"log"
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

func init() {
	log.Printf("Embedded Templatee Mode")
}

func GetCompiledSiteTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	f, err := fs.Sub(siteTemplatesFS, "site")
	if err != nil {
		panic(err)
	}
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(f, "*.gohtml", "*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	f, err := fs.Sub(notificationTemplatesFS, "notifications")
	if err != nil {
		panic(err)
	}
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(f, "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	f, err := fs.Sub(emailHtmlTemplatesFS, "email")
	if err != nil {
		panic(err)
	}
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(f, "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	f, err := fs.Sub(emailTextTemplatesFS, "email")
	if err != nil {
		panic(err)
	}
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(f, "*.gotxt"))
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
