package templates

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	htemplate "html/template"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	ttemplate "text/template"
	"time"

	"github.com/arran4/goa4web/core/consts"
)

// embeddedFS contains site templates, notification templates, email templates and static assets.
//
//go:embed site/*.gohtml site/*/*.gohtml notifications/*.gotxt email/*.gohtml email/*.gotxt assets/*
var embeddedFS embed.FS

var (
	// templateDir allows overriding the embedded templates for development.
	templateDir     string
	templateDirOnce sync.Once
)

// SetDir sets the directory to load templates from. This is intended for testing and development.
func SetDir(dir string) {
	templateDirOnce.Do(func() {
		templateDir = dir
	})
}

type config struct {
	Dir string
}

type Option func(*config)

func WithDir(dir string) Option {
	return func(c *config) {
		c.Dir = dir
	}
}

// newCfg creates a new config, applying the package-level templateDir if set.
func newCfg(opts ...Option) *config {
	cfg := &config{
		Dir: templateDir,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

var (
	assetHashes     = map[string]string{}
	assetHashesLock sync.RWMutex
	siteTemplates   *htemplate.Template
	siteTemplatesMu sync.Mutex
	templatesDir    string // To override embedded fs for tests/dev
)

// Asset reads an asset file from the configured source (embedded or local).
func Asset(name string) ([]byte, error) {
	if templatesDir != "" {
		return os.ReadFile(filepath.Join(templatesDir, "assets", name))
	}
	return embeddedFS.ReadFile("assets/" + name)
}

func init() {
	// Pre-compute hashes for embedded assets to avoid runtime overhead in production
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

func GetAssetHash(webPath string, opts ...Option) string {
	cfg := newCfg(opts...)
	base := path.Base(webPath)

	// If in development mode (serving from local directory), always recompute to reflect changes immediately.
	if cfg.Dir != "" {
		b, err := getAssetContent(base, cfg)
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

	b, err := getAssetContent(base, cfg)
	if err != nil {
		return webPath
	}

	sum := sha256.Sum256(b)
	h = hex.EncodeToString(sum[:])[:16]
	assetHashes[base] = h
	return webPath + "?v=" + h
}

func getAssetContent(name string, cfg *config) ([]byte, error) {
	if cfg.Dir == "" {
		return embeddedFS.ReadFile("assets/" + name)
	}
	return os.ReadFile(filepath.Join(cfg.Dir, "assets", name))
}

func getFS(sub string, cfg *config) fs.FS {
	if cfg.Dir == "" {
		log.Println("Embedded Template Mode")
		f, err := fs.Sub(embeddedFS, sub)
		if err != nil {
			panic(err)
		}
		return f
	}
	return os.DirFS(filepath.Join(cfg.Dir, sub))
}

func readFile(name string, opts ...Option) []byte {
	cfg := newCfg(opts...)
	if cfg.Dir == "" {
		b, err := embeddedFS.ReadFile(name)
		if err != nil {
			panic(err)
		}
		return b
	}
	b, err := os.ReadFile(filepath.Join(cfg.Dir, name))
	if err != nil {
		panic(err)
	}
	return b
}

func GetCompiledSiteTemplates(funcs htemplate.FuncMap, opts ...Option) *htemplate.Template {
	cfg := newCfg(opts...)

	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	funcs["assetHash"] = func(p string) string {
		return GetAssetHash(p, opts...)
	}
	funcs["url"] = func(s string) htemplate.URL {
		return htemplate.URL(s)
	}

	// Try to use cached templates if we are using embedded assets (no custom directory)
	if cfg.Dir == "" {
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
		// NOTE: Optimistic locking or double check could be better, but parsing is fast enough.
	}

	fsys := getFS("site", cfg)

	// Create root template.
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
		if cfg.Dir == "" {
			siteTemplatesMu.Unlock()
		}
		panic(err)
	}

	if cfg.Dir == "" {
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

func GetCompiledNotificationTemplates(funcs ttemplate.FuncMap, opts ...Option) *ttemplate.Template {
	cfg := newCfg(opts...)
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("notifications", cfg), "*.gotxt"))
}

func GetCompiledEmailHtmlTemplates(funcs htemplate.FuncMap, opts ...Option) *htemplate.Template {
	cfg := newCfg(opts...)
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
	return htemplate.Must(htemplate.New("").Funcs(funcs).ParseFS(getFS("email", cfg), "*.gohtml"))
}

func GetCompiledEmailTextTemplates(funcs ttemplate.FuncMap, opts ...Option) *ttemplate.Template {
	cfg := newCfg(opts...)
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
	return ttemplate.Must(ttemplate.New("").Funcs(funcs).ParseFS(getFS("email", cfg), "*.gotxt"))
}

func GetMainCSSData(opts ...Option) []byte { return readFile("assets/main.css", opts...) }

// GetFaviconData returns the site's favicon image.
func GetFaviconData(opts ...Option) []byte { return readFile("assets/favicon.svg", opts...) }

// GetFaviconPNG returns the site's favicon image as PNG.
func GetFaviconPNG(opts ...Option) []byte { return readFile("assets/favicon.png", opts...) }

// GetPasteImageJSData returns the JavaScript that enables image pasting.
func GetPasteImageJSData(opts ...Option) []byte { return readFile("assets/pasteimg.js", opts...) }

// GetNotificationsJSData returns the JavaScript used for real-time notification updates.
func GetNotificationsJSData(opts ...Option) []byte {
	return readFile("assets/notifications.js", opts...)
}

// GetRoleGrantsEditorJSData returns the JavaScript powering the role grants drag-and-drop editor.
func GetRoleGrantsEditorJSData(opts ...Option) []byte {
	return readFile("assets/role_grants_editor.js", opts...)
}

// GetPrivateForumJSData returns the JavaScript for private forum pages.
func GetPrivateForumJSData(opts ...Option) []byte {
	return readFile("assets/private_forum.js", opts...)
}

// GetTopicLabelsJSData returns the JavaScript for topic label editing.
func GetTopicLabelsJSData(opts ...Option) []byte { return readFile("assets/topic_labels.js", opts...) }

// GetSiteJSData returns the main site JavaScript.
func GetSiteJSData(opts ...Option) []byte { return readFile("assets/site.js", opts...) }

// GetA4CodeJSData returns the A4Code parser/converter JavaScript.
func GetA4CodeJSData(opts ...Option) []byte { return readFile("assets/a4code.js", opts...) }

// GetRobotsTXTData returns the robots.txt file.
func GetRobotsTXTData(opts ...Option) []byte { return readFile("assets/robots.txt", opts...) }

// ListSiteTemplateNames returns the relative paths of all site templates
// (under the site/ directory), e.g. "news/postPage.gohtml".
func ListSiteTemplateNames(opts ...Option) []string {
	cfg := newCfg(opts...)
	var names []string
	fsys := getFS("site", cfg)
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
func TemplateExists(name string, opts ...Option) bool {
	cfg := newCfg(opts...)
	fsys := getFS("site", cfg)
	if _, err := fs.Stat(fsys, name); err == nil {
		return true
	}
	return false
}

// EmailTemplateExists reports whether an email template with the given relative path
// exists in the current template source (embedded or templatesDir).
func EmailTemplateExists(name string, opts ...Option) bool {
	cfg := newCfg(opts...)
	fsys := getFS("email", cfg)
	if _, err := fs.Stat(fsys, name); err == nil {
		return true
	}
	return false
}

// NotificationTemplateExists reports whether a notification template with the given relative path
// exists in the current template source (embedded or templatesDir).
func NotificationTemplateExists(name string, opts ...Option) bool {
	cfg := newCfg(opts...)
	fsys := getFS("notifications", cfg)
	if _, err := fs.Stat(fsys, name); err == nil {
		return true
	}
	return false
}
