package templates

import _ "embed"

// UpdateEmailText contains the default notification email template.
//
//go:embed templates/updateEmail.txt
var UpdateEmailText string
