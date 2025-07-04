package templates

import _ "embed"

// BlogEmailText contains the new blog post notification email template.
//
//go:embed templates/blogEmail.txt
var BlogEmailText string
