package runtimeconfig

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

// extendedUsageFS embeds templates providing additional help text for configuration options.
//
//go:embed templates/*.txt
var extendedUsageFS embed.FS

// ExtendedUsage reads the named template and returns its contents as a string.
// The empty string is returned when the template does not exist or the name is blank.
// ExtendedUsage reads the named template and returns its contents rendered with
// information from the database driver registry. When the template does not
// exist or the name is blank an empty string is returned.
func ExtendedUsage(name string) string {
	if name == "" {
		return ""
	}
	b, err := extendedUsageFS.ReadFile("templates/" + name)
	if err != nil {
		return ""
	}
	tmpl, err := template.New(name).Parse(string(b))
	if err != nil {
		return ""
	}
	data := struct {
		Drivers []dbdrivers.DBDriver
	}{Drivers: dbdrivers.Registry}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}
