package templates

import _ "embed"

// SignupEmailHTML contains the HTML admin signup notification template.
//
//go:embed templates/signupEmail.html
var SignupEmailHTML string
