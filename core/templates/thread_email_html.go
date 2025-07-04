package templates

import _ "embed"

// ThreadEmailHTML contains the HTML new thread notification email template.
//
//go:embed templates/threadEmail.html
var ThreadEmailHTML string
