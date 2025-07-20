package notifications

import "github.com/arran4/goa4web/internal/eventbus"

type EmailTemplates struct {
	Text    string
	HTML    string
	Subject string
}

// NewEmailTemplates returns EmailTemplates populated with file names derived
// from prefix.
func NewEmailTemplates(prefix string) *EmailTemplates {
	return &EmailTemplates{
		Text:    EmailTextTemplateFilenameGenerator(prefix),
		HTML:    EmailHTMLTemplateFilenameGenerator(prefix),
		Subject: EmailSubjectTemplateFilenameGenerator(prefix),
	}
}

// AdminEmailTemplateProvider indicates the notification should be sent via
// email to administrators using the provided templates.
type AdminEmailTemplateProvider interface {
	AdminEmailTemplate() *EmailTemplates
	AdminInternalNotificationTemplate() *string
}

// SelfNotificationTemplateProvider is used for mandatory self notifications such as password
// resets or verifications.
type SelfNotificationTemplateProvider interface {
	SelfEmailTemplate() *EmailTemplates
	SelfInternalNotificationTemplate() *string
}

// DirectEmailNotificationTemplateProvider specifies templates for an email sent
// directly to an address independent of the user's primary email.
// The address itself is obtained from the event data via DirectEmailAddress.
// Internal notifications are not supported for this provider.
type DirectEmailNotificationTemplateProvider interface {
	DirectEmailAddress(evt eventbus.Event) string
	DirectEmailTemplate() *EmailTemplates
}

// SubscribersNotificationTemplateProvider indicates the notification should be delivered to
// subscribed users.
type SubscribersNotificationTemplateProvider interface {
	SubscribedEmailTemplate() *EmailTemplates
	SubscribedInternalNotificationTemplate() *string
}

// AutoSubscribeProvider describes events that automatically create a
// subscription when user preferences allow.
type AutoSubscribeProvider interface {
	// AutoSubscribePath returns the action name and URI used when creating the
	// subscription. The event may provide additional context required to build
	// the path.
	AutoSubscribePath(evt eventbus.TaskEvent) (string, string)
}

// TargetUsersNotificationProvider indicates the notification should be delivered
// to the returned user IDs.
type TargetUsersNotificationProvider interface {
	TargetUserIDs(evt eventbus.TaskEvent) []int32
	TargetEmailTemplate() *EmailTemplates
	TargetInternalNotificationTemplate() *string
}
