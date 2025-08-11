//go:build live
// +build live

package templates

import (
	htemplate "html/template"
	"log"
	"os"
	ttemplate "text/template"
	"time"
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
	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(os.DirFS("core/templates/email"), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	if funcs == nil {
		funcs = ttemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
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

// GetPrivateForumJSData returns the JavaScript for private forum pages.
func GetPrivateForumJSData() []byte {
	b, err := os.ReadFile("core/templates/assets/private_forum.js")
	if err != nil {
		panic(err)
	}
	return b
}

// GetTopicLabelsJSData returns the JavaScript for topic label editing.
func GetTopicLabelsJSData() []byte {
	b, err := os.ReadFile("core/templates/assets/topic_labels.js")
	if err != nil {
		panic(err)
	}
	return b
}
