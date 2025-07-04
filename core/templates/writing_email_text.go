package templates

import _ "embed"

// WritingEmailText contains the new writing notification email template.
//
//go:embed templates/writingEmail.txt
var WritingEmailText string
