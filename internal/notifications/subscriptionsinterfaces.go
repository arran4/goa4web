package notifications

import (
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

type EmailTemplates struct {
	Text    string
	HTML    string
	Subject string
}

// EmailTemplateName is a strongly-typed name for email templates (prefix).
type EmailTemplateName string

func (e EmailTemplateName) String() string {
	return string(e)
}

func (e EmailTemplateName) EmailTemplates() *EmailTemplates {
	return NewEmailTemplates(string(e))
}

func (e EmailTemplateName) NotificationTemplate() string {
	return NotificationTemplateFilenameGenerator(string(e))
}

func (e EmailTemplateName) RequiredTemplates() []tasks.Template {
	et := e.EmailTemplates()
	return []tasks.Template{
		tasks.Template(et.Text),
		tasks.Template(et.HTML),
		tasks.Template(et.Subject),
		tasks.Template(e.NotificationTemplate()),
	}
}

// NotificationTemplateName is a strongly-typed name for internal notification templates.
type NotificationTemplateName string

func (n NotificationTemplateName) String() string {
	return string(n)
}

func (n NotificationTemplateName) NotificationTemplate() string {
	return NotificationTemplateFilenameGenerator(string(n))
}

func (n NotificationTemplateName) RequiredTemplates() []tasks.Template {
	return []tasks.Template{
		tasks.Template(n.NotificationTemplate()),
	}
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
	AdminEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool)
	AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string
}

// SelfNotificationTemplateProvider is used for mandatory self notifications such as password
// resets or verifications.
type SelfNotificationTemplateProvider interface {
	SelfEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool)
	SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string
}

// SelfEmailBroadcaster indicates the notification should be sent to all
// verified email addresses of the user instead of only the highest priority.
type SelfEmailBroadcaster interface {
	SelfEmailBroadcast() bool
}

// DirectEmailNotificationTemplateProvider specifies templates for an email sent
// directly to an address independent of the user's primary email.
// The address itself is obtained from the event data via DirectEmailAddress.
// Internal notifications are not supported for this provider.
type DirectEmailNotificationTemplateProvider interface {
	DirectEmailAddress(evt eventbus.TaskEvent) (string, error)
	DirectEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool)
}

// SubscribersNotificationTemplateProvider indicates the notification should be delivered to
// subscribed users.
type SubscribersNotificationTemplateProvider interface {
	SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool)
	SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string
}

// AutoSubscribeProvider describes events that automatically create a
// subscription when user preferences allow.
type AutoSubscribeProvider interface {
	// AutoSubscribePath returns the action name and URI used when creating the
	// subscription. The event may provide additional context required to build
	// the path.
	AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error)
}

// TargetUsersNotificationProvider indicates the notification should be delivered
// to the returned user IDs.
type TargetUsersNotificationProvider interface {
	TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error)
	TargetEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool)
	TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string
}

// GrantsRequiredProvider exposes the permission context for subscription
// notifications. Implementations return one or more GrantRequirement values
// checked with `SystemCheckGrant` before delivering a message.
type GrantsRequiredProvider interface {
	GrantsRequired(evt eventbus.TaskEvent) ([]GrantRequirement, error)
}
