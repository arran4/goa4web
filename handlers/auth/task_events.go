package auth

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterTask represents user registration.
var RegisterTask = hcommon.NewTaskEvent(hcommon.TaskRegister)

// LoginTask represents user login.
var LoginTask = hcommon.NewTaskEvent(hcommon.TaskLogin)
