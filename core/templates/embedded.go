//go:build !live
// +build !live

package templates

import (
	"embed"
	htemplate "html/template"
	"io/fs"
	"log"
	ttemplate "text/template"
	"time"
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
	//go:embed "assets/role_grants_editor.js"
	roleGrantsEditorJSData []byte
	//go:embed "assets/private_forum.js"
	privateForumJSData []byte
	//go:embed "assets/topic_labels.js"
	topicLabelsJSData []byte
)

func init() {
	log.Printf("Embedded Template Mode")
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
	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
	f, err := fs.Sub(emailHtmlTemplatesFS, "email")
	if err != nil {
		panic(err)
	}
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(f, "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	if funcs == nil {
		funcs = ttemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
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

// GetRoleGrantsEditorJSData returns the JavaScript powering the
// role grants drag-and-drop editor.
func GetRoleGrantsEditorJSData() []byte { return roleGrantsEditorJSData }

// GetPrivateForumJSData returns the JavaScript for private forum pages.
func GetPrivateForumJSData() []byte { return privateForumJSData }

// GetTopicLabelsJSData returns the JavaScript for topic label editing.
func GetTopicLabelsJSData() []byte { return topicLabelsJSData }
