package templates

import _ "embed"

// PasswordResetEmailText is the default password reset email template.
//
//go:embed templates/passwordResetEmail.txt
var PasswordResetEmailText string
