package notifications

type EmailTemplates struct {
	Text    string
	HTML    string
	Subject string
}

// CreateEmail renders the templates using emailAddr and data.
// This is a convenience wrapper around RenderEmailFromTemplates.
func (et *EmailTemplates) CreateEmail(emailAddr string, data interface{}) ([]byte, error) {
	return RenderEmailFromTemplates(emailAddr, et, data)
}

// NewEmailTemplates returns EmailTemplates populated with file names derived
// from prefix.
func NewEmailTemplates(prefix string) *EmailTemplates {
	return &EmailTemplates{
		Text:    prefix + ".gotxt",
		HTML:    prefix + ".gohtml",
		Subject: prefix + "Subject.gotxt",
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
	// subscription.
	AutoSubscribePath() (string, string)
}
