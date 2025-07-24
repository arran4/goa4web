package config

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

//go:embed templates/*.txt
var tmplFS embed.FS

// ExtendedUsage renders the named template with data from dbdrivers.Registry.
func ExtendedUsage(name string, reg *dbdrivers.Registry) (string, error) {
	b, err := tmplFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	t, err := template.New(name).Parse(string(b))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, reg); err != nil {
		return "", err
	}
	return buf.String(), nil
}
