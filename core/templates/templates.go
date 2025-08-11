package templates

import (
	"embed"
	htemplate "html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	ttemplate "text/template"
	"time"
)

// templatesDir holds the optional directory to load templates from.
var templatesDir string

// embeddedFS contains site templates, notification templates, email templates and static assets.
//
//go:embed site/*.gohtml site/*/*.gohtml notifications/*.gotxt email/*.gohtml email/*.gotxt assets/*
var embeddedFS embed.FS

// SetDir configures templates to be loaded from dir. When dir is empty the embedded templates are used.
func SetDir(dir string) {
	templatesDir = dir
	if dir == "" {
		log.Printf("Embedded Template Mode")
	} else {
		log.Printf("Live Template Mode: %s", dir)
	}
}

func getFS(sub string) fs.FS {
	if templatesDir == "" {
		f, err := fs.Sub(embeddedFS, sub)
		if err != nil {
			panic(err)
		}
		return f
	}
	return os.DirFS(filepath.Join(templatesDir, sub))
}

func readFile(name string) []byte {
	if templatesDir == "" {
		b, err := embeddedFS.ReadFile(name)
		if err != nil {
			panic(err)
		}
		return b
	}
	b, err := os.ReadFile(filepath.Join(templatesDir, name))
	if err != nil {
		panic(err)
	}
	return b
}

func GetCompiledSiteTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(getFS("site"), "*.gohtml", "*/*.gohtml"))
}

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("notifications"), "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap) *htemplate.Template {
	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(getFS("email"), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap) *ttemplate.Template {
	if funcs == nil {
		funcs = ttemplate.FuncMap{}
	}
	if _, ok := funcs["localTime"]; !ok {
		funcs["localTime"] = func(t time.Time) time.Time { return t }
	}
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("email"), "*.gotxt"))
}

func GetMainCSSData() []byte { return readFile("assets/main.css") }

// GetFaviconData returns the site's favicon image data.
func GetFaviconData() []byte { return readFile("assets/favicon.svg") }

// GetPasteImageJSData returns the JavaScript that enables image pasting.
func GetPasteImageJSData() []byte { return readFile("assets/pasteimg.js") }

// GetNotificationsJSData returns the JavaScript used for real-time notification updates.
func GetNotificationsJSData() []byte { return readFile("assets/notifications.js") }

// GetRoleGrantsEditorJSData returns the JavaScript powering the role grants drag-and-drop editor.
func GetRoleGrantsEditorJSData() []byte { return readFile("assets/role_grants_editor.js") }

// GetPrivateForumJSData returns the JavaScript for private forum pages.
func GetPrivateForumJSData() []byte { return readFile("assets/private_forum.js") }

// GetTopicLabelsJSData returns the JavaScript for topic label editing.
func GetTopicLabelsJSData() []byte { return readFile("assets/topic_labels.js") }
