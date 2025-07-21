package blogs

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	db "github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	notif "github.com/arran4/goa4web/internal/notifications"
)

// AddBlogTask encapsulates creating a blog entry.
type AddBlogTask struct{ tasks.TaskString }

var addBlogTask = &AddBlogTask{TaskString: TaskAdd}

var _ tasks.Task = (*AddBlogTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*AddBlogTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AddBlogTask)(nil)
var _ notif.GrantsRequiredProvider = (*AddBlogTask)(nil)

func (AddBlogTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogAddEmail")
}

func (AddBlogTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogAddEmail")
	return &v
}

func (AddBlogTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("blogAddEmail")
}

func (AddBlogTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("blog_add")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for new blog entries.
func (AddBlogTask) GrantsRequired(evt eventbus.TaskEvent) []notif.GrantRequirement {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "blogs", Item: "entry", ItemID: t.ID, Action: "view"}}
	}
	return nil
}

func (AddBlogTask) Page(w http.ResponseWriter, r *http.Request)   { BlogAddPage(w, r) }
func (AddBlogTask) Action(w http.ResponseWriter, r *http.Request) { BlogAddActionPage(w, r) }

func BlogAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Mode               string
	}

	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(config.AppRuntimeConfig.DefaultLanguage)),
		Mode:               "Add",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "blogAddPage.gohtml", data)
}

func BlogAddActionPage(w http.ResponseWriter, r *http.Request) {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	id, err := queries.CreateBlogEntry(r.Context(), db.CreateBlogEntryParams{
		UsersIdusers:       uid,
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
			evt.Data["PostURL"] = cd.AbsoluteURL(fmt.Sprintf("/blogs/blog/%d", id))
			evt.Data["target"] = notif.Target{Type: "blog", ID: int32(id)}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", id), http.StatusTemporaryRedirect)
}
