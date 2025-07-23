package languages

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RenameLanguageTask performs a language rename action.
type RenameLanguageTask struct{ tasks.TaskString }

var renameLanguageTask = &RenameLanguageTask{TaskString: tasks.TaskString("Rename Language")}

func (RenameLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cname := r.PostFormValue("cname")
	if err := queries.RenameLanguage(r.Context(), db.RenameLanguageParams{
		Nameof:     sql.NullString{Valid: true, String: cname},
		Idlanguage: int32(cid),
	}); err != nil {
		return fmt.Errorf("rename language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["LanguageID"] = cid
			evt.Data["LanguageName"] = cname
		}
	}
	return nil
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

func (DeleteLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var name string
	if rows, err := queries.FetchLanguages(r.Context()); err == nil {
		for _, l := range rows {
			if l.Idlanguage == int32(cid) {
				name = l.Nameof.String
				break
			}
		}
	}
	if err := queries.DeleteLanguage(r.Context(), int32(cid)); err != nil {
		return fmt.Errorf("delete language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["LanguageID"] = cid
			evt.Data["LanguageName"] = name
		}
	}
	return nil
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

func (CreateLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cname := r.PostFormValue("cname")
	res, err := queries.InsertLanguage(r.Context(), sql.NullString{String: cname, Valid: true})
	if err != nil {
		return fmt.Errorf("create language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if id, err := res.LastInsertId(); err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["LanguageID"] = id
				evt.Data["LanguageName"] = cname
			}
		}
	}
	return nil
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
