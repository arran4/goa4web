package templates

import _ "embed"

// SignupEmailText contains the admin user signup notification template.
//
//go:embed templates/signupEmail.txt
var SignupEmailText string
