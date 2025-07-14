package templates

import _ "embed"

// UserRejectedEmailHTML contains the HTML user rejection notification template.
//
//go:embed templates/userRejectedEmail.html
var UserRejectedEmailHTML string
