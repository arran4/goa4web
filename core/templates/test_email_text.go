package templates

import _ "embed"

// TestEmailText contains the test email template.
//
//go:embed templates/testEmail.txt
var TestEmailText string
