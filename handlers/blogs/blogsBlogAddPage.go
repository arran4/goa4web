package blogs

import (
	"database/sql"
	"fmt"
	"time"

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

func (AddBlogTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationBlogAddEmail"), true
}

func (AddBlogTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogAddEmail")
	return &v
}

func (AddBlogTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("blogAddEmail"), true
}

func (AddBlogTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := cd.ValidateCodeImagesForUser(uid, text); err != nil {
		return fmt.Errorf("validate images: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	id, err := queries.CreateBlogEntryForWriter(r.Context(), db.CreateBlogEntryForWriterParams{
		UsersIdusers: uid,
		LanguageID:   sql.NullInt32{Int32: int32(languageId), Valid: languageId != 0},
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
		Written:  time.Now().UTC(),
		Timezone: sql.NullString{String: cd.Location().String(), Valid: true},
		UserID:   sql.NullInt32{Int32: uid, Valid: true},
		ListerID: uid,
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
	cd.PageTitle = "Add Blog"
	if !(cd.IsAdmin() || cd.HasGrant("blogs", "entry", "post", 0)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	type Data struct {
		Languages          []*db.Language
		SelectedLanguageId int
		Mode               string
	}

	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		Mode:               "Add",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "blogAddPage.gohtml", data)
}
