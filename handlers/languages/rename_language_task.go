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

var _ tasks.Task = (*RenameLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RenameLanguageTask)(nil)

func (RenameLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	queries := cd.Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cname := r.PostFormValue("cname")
	if err := queries.AdminRenameLanguage(r.Context(), db.AdminRenameLanguageParams{
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
