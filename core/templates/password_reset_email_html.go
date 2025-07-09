package templates

import _ "embed"

// PasswordResetEmailHTML is the default HTML password reset email template.
//
//go:embed templates/passwordResetEmail.html
var PasswordResetEmailHTML string
