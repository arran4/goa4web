package dbstart

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	// schemaMismatchTmpl is the CLI message shown when the database schema version is unexpected.
	//
	//go:embed templates/schema_mismatch.txt
	schemaMismatchTmpl string
)

// RenderSchemaMismatch returns the formatted schema mismatch message.
func RenderSchemaMismatch(actual, expected int) string {
	exe := filepath.Base(os.Args[0])
	if !strings.HasSuffix(exe, "-admin") {
		exe += "-admin"
	}
	var buf bytes.Buffer
	tmpl := template.Must(template.New("schema").Parse(schemaMismatchTmpl))
	if err := tmpl.Execute(&buf, struct {
		Actual, Expected int
		Exe              string
	}{actual, expected, exe}); err != nil {
		log.Printf("schema mismatch template execute: %v", err)
	}
	return strings.TrimSpace(buf.String())
}
