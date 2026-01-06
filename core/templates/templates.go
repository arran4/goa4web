package templates

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	htemplate "html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	ttemplate "text/template"
	"text/template/parse"
	"time"

	"github.com/arran4/goa4web/core/consts"
)

// embeddedFS contains site templates, notification templates, email templates and static assets.
//
//go:embed site/*.gohtml site/*/*.gohtml notifications/*.gotxt email/*.gohtml email/*.gotxt assets/*
var embeddedFS embed.FS

// SetDir is deprecated and will panic.
func SetDir(dir string) {
	panic("SetDir is deprecated. Use functional options or arguments to load templates from a specific directory.")
}

var (
	assetHashes     = map[string]string{}
	assetHashesLock sync.RWMutex
	siteTemplates   *htemplate.Template
	siteTemplatesMu sync.Mutex
)

func init() {
	// Pre-compute hashes for embedded assets
	entries, err := embeddedFS.ReadDir("assets")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				b, err := embeddedFS.ReadFile("assets/" + e.Name())
				if err == nil {
					sum := sha256.Sum256(b)
					assetHashes[e.Name()] = hex.EncodeToString(sum[:])[:16]
				}
			}
		}
	}
}

func parseOptions(ops ...any) (string, error) {
	var dir string
	for _, op := range ops {
		switch v := op.(type) {
		case string:
			dir = v
		default:
			return "", fmt.Errorf("unknown option type %T", op)
		}
	}
	return dir, nil
}

func GetAssetHash(webPath string, ops ...any) string {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}

	base := path.Base(webPath)

	// If in development mode (serving from local directory), always recompute to reflect changes immediately.
	if dir != "" {
		b, err := getAssetContent(base, dir)
		if err != nil {
			return webPath
		}
		sum := sha256.Sum256(b)
		h := hex.EncodeToString(sum[:])[:16]
		return webPath + "?v=" + h
	}

	assetHashesLock.RLock()
	h, ok := assetHashes[base]
	assetHashesLock.RUnlock()
	if ok {
		return webPath + "?v=" + h
	}

	assetHashesLock.Lock()
	defer assetHashesLock.Unlock()
	if h, ok := assetHashes[base]; ok {
		return webPath + "?v=" + h
	}

	b, err := getAssetContent(base, "")
	if err != nil {
		return webPath
	}

	sum := sha256.Sum256(b)
	h = hex.EncodeToString(sum[:])[:16]
	assetHashes[base] = h
	return webPath + "?v=" + h
}

func getAssetContent(name string, dir string) ([]byte, error) {
	if dir == "" {
		return embeddedFS.ReadFile("assets/" + name)
	}
	return os.ReadFile(filepath.Join(dir, "assets", name))
}

func getFS(sub string, dir string) fs.FS {
	if dir == "" {
		f, err := fs.Sub(embeddedFS, sub)
		if err != nil {
			panic(err)
		}
		return f
	}
	return os.DirFS(filepath.Join(dir, sub))
}

func readFile(name string, ops ...any) []byte {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
	if dir == "" {
		b, err := embeddedFS.ReadFile(name)
		if err != nil {
			panic(err)
		}
		return b
	}
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		panic(err)
	}
	return b
}

func GetCompiledSiteTemplates(funcs htemplate.FuncMap, ops ...any) *htemplate.Template {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}

	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	funcs["assetHash"] = func(p string) string {
		return GetAssetHash(p, ops...)
	}

	// Try to use cached templates if we are using embedded assets (no custom directory)
	if dir == "" {
		siteTemplatesMu.Lock()
		if siteTemplates != nil {
			// Clone the cached template so the caller can execute it without
			// affecting the master copy (which would mark it as executed).
			t, err := siteTemplates.Clone()
			siteTemplatesMu.Unlock()
			if err != nil {
				panic(err)
			}
			return t.Funcs(funcs)
		}
		// If cache is missing, we must parse. We hold the lock to update cache.
	}

	fsys := getFS("site", dir)

	root := htemplate.New("root").Funcs(funcs)

	// Walk the sub-FS and parse every *.gohtml, naming templates by relative path.
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
		if dir == "" {
			siteTemplatesMu.Unlock()
		}
		panic(err)
	}

	if dir == "" {
		// Cache the unexecuted root
		siteTemplates = root
		siteTemplatesMu.Unlock()
	}

	// Always return a clone to the caller to ensure the root (cached or not)
	// remains unexecuted.
	t, err := root.Clone()
	if err != nil {
		panic(err)
	}
	return t
}

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap, ops ...any) *ttemplate.Template {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("notifications", dir), "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap, ops ...any) *htemplate.Template {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
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
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(getFS("email", dir), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap, ops ...any) *ttemplate.Template {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
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
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("email", dir), "*.gotxt"))
}

func GetMainCSSData(ops ...any) []byte { return readFile("assets/main.css", ops...) }

// GetFaviconData returns the site's favicon image data.
func GetFaviconData(ops ...any) []byte { return readFile("assets/favicon.svg", ops...) }

// GetPasteImageJSData returns the JavaScript that enables image pasting.
func GetPasteImageJSData(ops ...any) []byte { return readFile("assets/pasteimg.js", ops...) }

// GetNotificationsJSData returns the JavaScript used for real-time notification updates.
func GetNotificationsJSData(ops ...any) []byte { return readFile("assets/notifications.js", ops...) }

// GetRoleGrantsEditorJSData returns the JavaScript powering the role grants drag-and-drop editor.
func GetRoleGrantsEditorJSData(ops ...any) []byte {
	return readFile("assets/role_grants_editor.js", ops...)
}

// GetPrivateForumJSData returns the JavaScript for private forum pages.
func GetPrivateForumJSData(ops ...any) []byte { return readFile("assets/private_forum.js", ops...) }

// GetTopicLabelsJSData returns the JavaScript for topic label editing.
func GetTopicLabelsJSData(ops ...any) []byte { return readFile("assets/topic_labels.js", ops...) }

// GetSiteJSData returns the main site JavaScript.
func GetSiteJSData(ops ...any) []byte { return readFile("assets/site.js", ops...) }

// GetA4CodeJSData returns the A4Code parser/converter JavaScript.
func GetA4CodeJSData(ops ...any) []byte { return readFile("assets/a4code.js", ops...) }

// ListSiteTemplateNames returns the relative paths of all site templates
// (under the site/ directory), e.g. "news/postPage.gohtml".
func ListSiteTemplateNames(ops ...any) []string {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
	var names []string
	fsys := getFS("site", dir)
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
func TemplateExists(name string, ops ...any) bool {
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
	fsys := getFS("site", dir)
	if _, err := fs.Stat(fsys, name); err == nil {
		return true
	}
	return false
}

// LoadAllTemplatesMap loads all site templates into the allTemplates map.
// Deprecated: Use ListSiteTemplateNames or TemplateExists
func LoadAllTemplatesMap(ops ...any) {
	// No-op or implementation if needed for backward compat but user wants to eliminate globals.
	// We can leave this empty or removed.
}

// IsTemplateAvailable checks if a template is available in the site templates.
// It checks for file existence first, then scans for defined templates.
func IsTemplateAvailable(name string, ops ...any) bool {
	// We handle parseOptions inside TemplateExists, call it wrapper if you want.
	dir, err := parseOptions(ops...)
	if err != nil {
		panic(err)
	}
	// Note: TemplateExists calls parseOptions too. We can pass the string dir directly if we recursively call IsTemplateAvailable?
	// But ops are any. If we pass string, parseOptions handles it.

	if TemplateExists(name, ops...) {
		return true
	}

	// Re-get fs using dir from options
	fsys := getFS("site", dir)

	found := false
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".gohtml" {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// Try parsing
		trees, err := parse.Parse(path, string(content), "{{", "}}")
		if err == nil {
			if _, ok := trees[name]; ok {
				found = true
				return fs.SkipAll
			}
		} else {
			// Fallback regex
			re := regexp.MustCompile(`{{\s*define\s+"([^"]+)"\s*}}`)
			matches := re.FindAllStringSubmatch(string(content), -1)
			for _, m := range matches {
				if m[1] == name {
					found = true
					return fs.SkipAll
				}
			}
		}
		return nil
	})

	return found
}
