package auth

import "github.com/arran4/goa4web/internal/tasks"

// TODO remove what is unused here

// TODO this should be the basis for all the others.

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (

	// TaskLogin performs a user login.
	TaskLogin tasks.TaskString = "Login"

	// TaskRegister registers a new user account.
	TaskRegister = "Register"

	// TaskUserResetPassword resets a user's password.
	TaskUserResetPassword tasks.TaskString = "Password Reset"

	// TaskPasswordVerify verifies a password reset code.
	TaskPasswordVerify = "Password Verify"
)
