package auth

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// RegisterTask represents user registration.
var RegisterTask = tasks.NewTaskEventWithHandlers(tasks.TaskRegister, RegisterActionPage)

// LoginTask represents user login.
var LoginTask = tasks.NewTaskEventWithHandlers(tasks.TaskLogin, LoginActionPage)

// VerifyPasswordTask handles password reset verification.
var VerifyPasswordTask = tasks.NewTaskEvent(tasks.TaskPasswordVerify)
