package templates

import _ "embed"

// TestEmailHTML contains the HTML version of the test email template.
//
//go:embed templates/testEmail.html
var TestEmailHTML string
