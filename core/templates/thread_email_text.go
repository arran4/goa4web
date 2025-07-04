package templates

import _ "embed"

// ThreadEmailText contains the new thread notification email template.
//
//go:embed templates/threadEmail.txt
var ThreadEmailText string
