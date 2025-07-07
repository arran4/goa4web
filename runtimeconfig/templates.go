package runtimeconfig

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

//go:embed templates/*.txt
var tmplFS embed.FS

// ExtendedUsage renders the named template with data from dbdrivers.Registry.
func ExtendedUsage(name string) (string, error) {
	b, err := tmplFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	t, err := template.New(name).Parse(string(b))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, dbdrivers.Registry); err != nil {
		return "", err
	}
	return buf.String(), nil
}
