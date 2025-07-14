package auth

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterTask represents user registration.
var RegisterTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskRegister, RegisterPage, RegisterActionPage)

// LoginTask represents user login.
var LoginTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskLogin, LoginUserPassPage, LoginActionPage)

// ResetPasswordTask handles password reset requests.
var ResetPasswordTask = hcommon.NewTaskEventWithHandlers(hcommon.TaskUserResetPassword, ForgotPasswordPage, ForgotPasswordActionPage)

// VerifyPasswordTask handles password reset verification.
var VerifyPasswordTask = hcommon.NewTaskEvent(hcommon.TaskPasswordVerify)
