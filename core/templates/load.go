package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
)

// LoadSiteTemplates parses all .gohtml templates under root using the provided
// FuncMap. It returns the parsed *template.Template or an error.
func LoadSiteTemplates(funcs template.FuncMap, root string) (*template.Template, error) {
	t := template.New("root").Funcs(funcs)
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
	return t.ParseFiles(files...)
}
