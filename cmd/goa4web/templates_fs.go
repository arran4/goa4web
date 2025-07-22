package main

import (
	"embed"
)

// templatesFS contains all CLI usage templates.
//
//go:embed templates/*.txt
var templatesFS embed.FS
