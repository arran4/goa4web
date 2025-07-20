package blogs

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	notif "github.com/arran4/goa4web/internal/notifications"
)

// EditBlogTask updates an existing blog entry.
type EditBlogTask struct{ tasks.TaskString }

var editBlogTask = &EditBlogTask{TaskString: TaskEdit}

var _ tasks.Task = (*EditBlogTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EditBlogTask)(nil)

func (EditBlogTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogEditEmail")
}

func (EditBlogTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogEditEmail")
	return &v
}

func (EditBlogTask) Page(w http.ResponseWriter, r *http.Request)   { BlogEditPage(w, r) }
func (EditBlogTask) Action(w http.ResponseWriter, r *http.Request) { BlogEditActionPage(w, r) }

func BlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		Blog               *db.GetBlogEntryForUserByIdRow
		SelectedLanguageId int
		Mode               string
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
		Mode:               "Edit",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	row := r.Context().Value(consts.KeyBlogEntry).(*db.GetBlogEntryForUserByIdRow)
	data.Blog = row

	handlers.TemplateHandler(w, r, "blogEditPage.gohtml", data)
}

func BlogEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	row := r.Context().Value(consts.KeyBlogEntry).(*db.GetBlogEntryForUserByIdRow)

	err = queries.UpdateBlogEntry(r.Context(), db.UpdateBlogEntryParams{
		Idblogs:            row.Idblogs,
		LanguageIdlanguage: int32(languageId),
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["PostURL"] = handlers.AbsoluteURL(r, fmt.Sprintf("/blogs/blog/%d", row.Idblogs))
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", row.Idblogs), http.StatusTemporaryRedirect)
}
