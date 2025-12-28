package templates

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadSiteTemplates parses all .gohtml templates under root using the provided
// FuncMap. It uses the local filesystem (os.DirFS) and is a convenience wrapper
// around LoadSiteTemplatesFS.
func LoadSiteTemplates(funcs template.FuncMap, root string) (*template.Template, error) {
	return LoadSiteTemplatesFS(funcs, os.DirFS(root), ".")
}

// LoadSiteTemplatesFS parses all .gohtml templates from the provided fs.FS.
// This allows tests to provide an in-memory fs.FS if they need to avoid
// touching the real filesystem.
func LoadSiteTemplatesFS(funcs template.FuncMap, fsys fs.FS, root string) (*template.Template, error) {
	t := template.New("root").Funcs(funcs)
	var files []string
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".gohtml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return t, nil
	}
	return t.ParseFS(fsys, files...)
}
