package languages

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RenameLanguageTask performs a language rename action.
type RenameLanguageTask struct{ tasks.TaskString }

var renameLanguageTask = &RenameLanguageTask{TaskString: tasks.TaskString("Rename Language")}

func (RenameLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesRenamePage(w, r)
}

func (RenameLanguageTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLanguageRenameEmail")
}

func (RenameLanguageTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLanguageRenameEmail")
	return &v
}

// DeleteLanguageTask removes a language entry.
type DeleteLanguageTask struct{ tasks.TaskString }

var deleteLanguageTask = &DeleteLanguageTask{TaskString: tasks.TaskString("Delete Language")}

func (DeleteLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesDeletePage(w, r)
}

func (DeleteLanguageTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLanguageDeleteEmail")
}

func (DeleteLanguageTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLanguageDeleteEmail")
	return &v
}

// CreateLanguageTask creates a new language.
type CreateLanguageTask struct{ tasks.TaskString }

var createLanguageTask = &CreateLanguageTask{TaskString: tasks.TaskString("Create Language")}

func (CreateLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesCreatePage(w, r)
}

func (CreateLanguageTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLanguageCreateEmail")
}

func (CreateLanguageTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLanguageCreateEmail")
	return &v
}

var (
	_ tasks.Task                       = (*RenameLanguageTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*RenameLanguageTask)(nil)
	_ tasks.Task                       = (*DeleteLanguageTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*DeleteLanguageTask)(nil)
	_ tasks.Task                       = (*CreateLanguageTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CreateLanguageTask)(nil)
)
