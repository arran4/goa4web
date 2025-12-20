package templates

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"sync"
	"text/template/parse"
)

var (
	allTemplates     map[string]struct{}
	allTemplatesOnce sync.Once
)

// LoadAllTemplatesMap builds a map of all available templates, including files and
// those defined within files using {{ define }}.
func LoadAllTemplatesMap() {
	allTemplatesOnce.Do(func() {
		allTemplates = make(map[string]struct{})
		fsys := getFS("site")

		err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".gohtml" {
				return nil
			}

			// Add the file path as a template name
			allTemplates[path] = struct{}{}

			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			// Parse to find defines
			// parse.Parse returns map[string]*Tree
			trees, err := parse.Parse(path, string(content), "{{", "}}")
			if err == nil {
				for name := range trees {
					allTemplates[name] = struct{}{}
				}
			} else {
				// Fallback to regex if parsing fails (e.g. due to missing funcs)
				re := regexp.MustCompile(`{{\s*define\s+"([^"]+)"\s*}}`)
				matches := re.FindAllStringSubmatch(string(content), -1)
				for _, m := range matches {
					allTemplates[m[1]] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	})
}

// IsTemplateAvailable returns true if the template exists as a file or is defined in one.
// It initializes the template map on first use.
func IsTemplateAvailable(name string) bool {
	LoadAllTemplatesMap()
	_, ok := allTemplates[name]
	return ok
}
