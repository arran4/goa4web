package templates

import (
	"embed"
	"github.com/arran4/goa4web/core/consts"
	htemplate "html/template"
	"io/fs"
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
func SetDir(dir string) { templatesDir = dir }

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
	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}

	fsys := getFS("site")

	root := htemplate.New("root").Funcs(funcs)

	// Walk the sub-FS and parse every *.gohtml, naming templates by relative path.
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".gohtml" {
			return nil
		}

		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// IMPORTANT: use path (the relative filename) as the template name.
		_, err = root.New(path).Parse(string(b))
		return err
	})
	if err != nil {
		panic(err)
	}

	return root
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
	if _, ok := funcs["formatLocalTime"]; !ok {
		funcs["formatLocalTime"] = func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format(consts.DisplayDateTimeFormat)
		}
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
	if _, ok := funcs["formatLocalTime"]; !ok {
		funcs["formatLocalTime"] = func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format(consts.DisplayDateTimeFormat)
		}
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

// GetSiteJSData returns the main site JavaScript.
func GetSiteJSData() []byte { return readFile("assets/site.js") }

// GetA4CodeJSData returns the A4Code parser/converter JavaScript.
func GetA4CodeJSData() []byte { return readFile("assets/a4code.js") }

// ListSiteTemplateNames returns the relative paths of all site templates
// (under the site/ directory), e.g. "news/postPage.gohtml".
func ListSiteTemplateNames() []string {
	var names []string
	fsys := getFS("site")
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".gohtml" {
			return nil
		}
		names = append(names, path)
		return nil
	})
	return names
}

// TemplateExists reports whether a site template with the given relative path
// exists in the current template source (embedded or templatesDir).
func TemplateExists(name string) bool {
	fsys := getFS("site")
	if _, err := fs.Stat(fsys, name); err == nil {
		return true
	}
	return false
}
