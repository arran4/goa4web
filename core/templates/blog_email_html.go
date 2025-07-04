package templates

import _ "embed"

// BlogEmailHTML contains the HTML blog post notification email template.
//
//go:embed templates/blogEmail.html
var BlogEmailHTML string
