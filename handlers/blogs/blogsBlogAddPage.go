package blogs

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"

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
func (AddBlogTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "blogs", Item: "entry", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

func (AddBlogTask) Page(w http.ResponseWriter, r *http.Request) { BlogAddPage(w, r) }
func (AddBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
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
		return fmt.Errorf("blog create fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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

	return handlers.RedirectHandler(fmt.Sprintf("/blogs/blog/%d", id))
}

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
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
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
