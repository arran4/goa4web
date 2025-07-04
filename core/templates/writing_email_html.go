package templates

import _ "embed"

// WritingEmailHTML contains the HTML writing notification email template.
//
//go:embed templates/writingEmail.html
var WritingEmailHTML string
