package user

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskSaveLanguages saves multiple languages at once.
	TaskSaveLanguages tasks.TaskString = "Save languages"
	// TaskSaveLanguage saves updates to a single language.
	TaskSaveLanguage tasks.TaskString = "Save language"
	// TaskSaveSize saves a user's paging preference.
	TaskSaveSize tasks.TaskString = "Save size"
	// TaskSavePublicProfile toggles a user's public profile setting.
	TaskSavePublicProfile tasks.TaskString = "Save public profile"
	// TaskSaveTimezone saves a user's timezone preference.
	TaskSaveTimezone tasks.TaskString = "Save timezone"
	// TaskSaveAppearance saves a user's custom CSS preference.
	TaskSaveAppearance tasks.TaskString = "Save appearance"
	// TaskSaveAll saves all changes in bulk.
	TaskSaveAll tasks.TaskString = "Save all"
	// TaskTestMail sends a test email to the current user.
	TaskTestMail tasks.TaskString = "Test mail"
	// TaskDismiss marks a notification as read.
	TaskDismiss tasks.TaskString = "Dismiss"

	// TaskResend attempts to resend a verification email.
	TaskResend tasks.TaskString = "Resend"

	// TaskAdd represents the "Add" action.
	TaskAdd tasks.TaskString = "Add"

	// TaskDelete removes an existing item.
	TaskDelete tasks.TaskString = "Delete"

	// TaskUpdate updates an existing item.
	TaskUpdate tasks.TaskString = "Update"

	// TaskUserAllow grants a user a role.
	TaskUserAllow tasks.TaskString = "User Allow"

	// TaskUserDisallow removes a user's role.
	TaskUserDisallow tasks.TaskString = "User Disallow"

	// TaskUserEmailVerification verifies a user's email address.
	TaskUserEmailVerification tasks.TaskString = "Email Verification"
)
