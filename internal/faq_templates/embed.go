package faq_templates

import (
	"embed"
	"io"
	"strings"
)

//go:embed *.txt
var templateFS embed.FS

// List returns the list of available template names.
func List() ([]string, error) {
	entries, err := templateFS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(e.Name(), ".txt"))
		}
	}
	return names, nil
}

// Get returns the content of a template by name.
func Get(name string) (string, error) {
	if !strings.HasSuffix(name, ".txt") {
		name += ".txt"
	}
	f, err := templateFS.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
