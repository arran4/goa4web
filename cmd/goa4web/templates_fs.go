package main

import (
	"embed"
	"fmt"
)

// templatesFS contains all CLI usage templates.
//
//go:embed templates/*.txt
var templatesFS embed.FS

// templateString returns the contents of the named template file.
func templateString(name string) string {
	b, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		panic(fmt.Errorf("read template %s: %w", name, err))
	}
	return string(b)
}
