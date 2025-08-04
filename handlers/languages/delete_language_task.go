package languages

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteLanguageTask removes a language entry.
type DeleteLanguageTask struct{ tasks.TaskString }

var deleteLanguageTask = &DeleteLanguageTask{TaskString: tasks.TaskString("Delete Language")}

var _ tasks.Task = (*DeleteLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*DeleteLanguageTask)(nil)

func (DeleteLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Forbidden"))
		})
	}
	queries := cd.Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var name string
	if rows, err := cd.Languages(); err == nil {
		for _, l := range rows {
			if l.Idlanguage == int32(cid) {
				name = l.Nameof.String
				break
			}
		}
	}
	if err := queries.AdminDeleteLanguage(r.Context(), int32(cid)); err != nil {
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
