package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"sync"
	"testing"
)

//go:embed site/*.gohtml site/*/*.gohtml
var testSiteTemplatesForRefs embed.FS

// TestBadTemplateReferences finds template {{define}} and {{template}} calls that don't exist.
func TestBadTemplateReferences(t *testing.T) {
	siteFS, err := fs.Sub(testSiteTemplatesForRefs, "site")
	if err != nil {
		t.Fatal(err)
	}

	templateActionRegex := regexp.MustCompile(`{{\s*-?\s*template\s*"([^"]+)"`)
	defineActionRegex := regexp.MustCompile(`{{\s*-?\s*define\s*"([^"]+)"`)

	// map[filename] -> []{called template name}
	calls := make(map[string][]string)
	// map[defined template name] -> filename
	definitions := make(map[string]string)
	// set of all template files, e.g. "index.gohtml", "admin/users.gohtml"
	templateFiles := make(map[string]struct{})

	var mu sync.Mutex

	err = fs.WalkDir(siteFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." || d.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}

		mu.Lock()
		templateFiles[path] = struct{}{}
		mu.Unlock()

		content, readErr := fs.ReadFile(siteFS, path)
		if readErr != nil {
			return readErr
		}
		scontent := string(content)

		// Find calls
		matches := templateActionRegex.FindAllStringSubmatch(scontent, -1)
		if len(matches) > 0 {
			mu.Lock()
			for _, match := range matches {
				calls[path] = append(calls[path], match[1])
			}
			mu.Unlock()
		}

		// Find definitions
		matches = defineActionRegex.FindAllStringSubmatch(scontent, -1)
		if len(matches) > 0 {
			mu.Lock()
			for _, match := range matches {
				if existingFile, ok := definitions[match[1]]; ok {
					t.Errorf("template '%s' defined in both %s and %s", match[1], existingFile, path)
				}
				definitions[match[1]] = path
			}
			mu.Unlock()
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk templates: %v", err)
	}

	// Now validate
	var badRefs []string
	for file, calledTemplates := range calls {
		for _, called := range calledTemplates {
			// A template call is valid if it refers to:
			// 1. A defined template (e.g. {{define "foo"}})
			_, isDefined := definitions[called]
			// 2. A file template (e.g. "index.gohtml")
			_, isFile := templateFiles[called]

			if !isDefined && !isFile {
				badRefs = append(badRefs, fmt.Sprintf("%s: calls non-existent template '%s'", file, called))
			}
		}
	}

	if len(badRefs) > 0 {
		t.Errorf("found bad template references:\n%s", strings.Join(badRefs, "\n"))
	}
}
