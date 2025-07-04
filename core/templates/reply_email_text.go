package templates

import _ "embed"

// ReplyEmailText contains the default reply notification email template.
//
//go:embed templates/replyEmail.txt
var ReplyEmailText string
