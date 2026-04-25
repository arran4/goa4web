package notifications

import "github.com/arran4/goa4web/internal/tasks"

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
