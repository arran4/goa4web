package templates

import (
	htemplate "html/template"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadSiteTemplates parses all .gohtml templates under root using the provided
// FuncMap. It uses the local filesystem (os.DirFS) and is a convenience wrapper
// around LoadSiteTemplatesFS.
func LoadSiteTemplates(funcs htemplate.FuncMap, root string) (*htemplate.Template, error) {
	return LoadSiteTemplatesFS(funcs, os.DirFS(root), ".")
}

// LoadSiteTemplatesFS parses all .gohtml templates from the provided fs.FS.
// This allows tests to provide an in-memory fs.FS if they need to avoid
// touching the real filesystem.
func LoadSiteTemplatesFS(funcs htemplate.FuncMap, fsys fs.FS, root string) (*htemplate.Template, error) {
	if funcs == nil {
		funcs = htemplate.FuncMap{}
	}
	// ensure assetHash helper is available just like GetCompiledSiteTemplates
	funcs["assetHash"] = GetAssetHash

	t := htemplate.New("root").Funcs(funcs)

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
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

		// name templates by their relative path to match embedded/template-dir behaviour
		_, err = t.New(path).Parse(string(b))
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}
