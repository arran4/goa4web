package templates

import _ "embed"

// UserRejectedEmailText contains the user rejection notification template.
//
//go:embed templates/userRejectedEmail.txt
var UserRejectedEmailText string
