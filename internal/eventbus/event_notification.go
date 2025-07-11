package eventbus

// EventNotification represents an event that should be dispatched to interested
// parties by the EventBus.
type EventNotification struct {
	// Source identifies the task that generated this notification.
	Source string
	// Path is a URI describing where the event originated from.
	Path string
	// UserID identifies the user responsible for the event.
	UserID int32
	// TemplateData holds arbitrary values passed to notification templates.
	TemplateData map[string]any
}

// AdminEmailTemplateProvider indicates the notification should be sent via
// email to administrators using the provided templates.
type AdminEmailTemplateProvider interface {
	AdminEmailTextTemplate() string
	AdminEmailHTMLTemplate() string
	BusMessage() *Bus
}

// AdminNotificationTemplateProvider indicates an admin view notification should
// be created using the provided template.
type AdminNotificationTemplateProvider interface {
	AdminNotificationTemplate() string
	BusMessage() *Bus
}

// SelfTemplateProvider is used for mandatory self notifications such as password
// resets or verifications.
type SelfTemplateProvider interface {
	SelfEmailTextTemplate() string
	SelfEmailHTMLTemplate() string
	SelfNotificationTemplate() string
	BusMessage() *Bus
}

// SubscribersTemplateProvider indicates the notification should be delivered to
// subscribed users.
type SubscribersTemplateProvider interface {
	SubscribedEmailTextTemplate() string
	SubscribedEmailHTMLTemplate() string
	SubscribedNotificationTemplate() string
	BusMessage() *Bus
}

// AutoSubscribeProvider describes events that automatically create a
// subscription when user preferences allow.
type AutoSubscribeProvider interface {
	// AutoSubscribePath returns the action name and URI used when creating the
	// subscription.
	AutoSubscribePath() (string, string)
}
