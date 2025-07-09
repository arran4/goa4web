package templates

import _ "embed"

// VerificationEmailText is the default verification email template.
//
//go:embed templates/verifyEmail.txt
var VerificationEmailText string
