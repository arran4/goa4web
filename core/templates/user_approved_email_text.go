package templates

import _ "embed"

// UserApprovedEmailText contains the user approval notification template.
//
//go:embed templates/userApprovedEmail.txt
var UserApprovedEmailText string
