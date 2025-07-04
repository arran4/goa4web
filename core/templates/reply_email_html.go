package templates

import _ "embed"

// ReplyEmailHTML contains the default HTML reply notification email template.
//
//go:embed templates/replyEmail.html
var ReplyEmailHTML string
