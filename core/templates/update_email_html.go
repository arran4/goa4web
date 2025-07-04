package templates

import _ "embed"

// UpdateEmailHTML contains the default HTML notification email template.
//
//go:embed templates/updateEmail.html
var UpdateEmailHTML string
