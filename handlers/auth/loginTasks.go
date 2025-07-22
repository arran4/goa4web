package auth

import "github.com/arran4/goa4web/internal/tasks"

// LoginTask handles rendering and processing of the login form.
type LoginTask struct {
	tasks.TaskString
}

var loginTask = &LoginTask{TaskString: TaskLogin}

// ensure LoginTask conforms to tasks.Task
var _ tasks.Task = (*LoginTask)(nil)

// VerifyPasswordTask verifies reset codes during login.
type VerifyPasswordTask struct {
	tasks.TaskString
}

var verifyPasswordTask = &VerifyPasswordTask{TaskString: TaskPasswordVerify}

// ensure VerifyPasswordTask conforms to tasks.Task
var _ tasks.Task = (*VerifyPasswordTask)(nil)
