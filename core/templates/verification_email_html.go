package templates

import _ "embed"

// VerificationEmailHTML is the default HTML verification email template.
//
//go:embed templates/verifyEmail.html
var VerificationEmailHTML string
