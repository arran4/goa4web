package templates

import _ "embed"

// UserApprovedEmailHTML contains the HTML user approval notification template.
//
//go:embed templates/userApprovedEmail.html
var UserApprovedEmailHTML string
