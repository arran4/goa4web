package main

import (
	emailpkg "github.com/arran4/goa4web/internal/email"
	emaildefaults "github.com/arran4/goa4web/internal/email/emaildefaults"
)

func init() { emaildefaults.Register(emailpkg.DefaultRegistry) }
