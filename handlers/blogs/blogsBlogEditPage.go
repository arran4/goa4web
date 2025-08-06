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
	"github.com/arran4/goa4web/internal/tasks"

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

func (EditBlogTask) Page(w http.ResponseWriter, r *http.Request) { BlogEditPage(w, r) }
func (EditBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	row := cd.CurrentBlogLoaded()

	if err = queries.UpdateBlogEntryForWriter(r.Context(), db.UpdateBlogEntryForWriterParams{
		EntryID:      row.Idblogs,
		GrantEntryID: sql.NullInt32{Int32: row.Idblogs, Valid: true},
		LanguageID:   int32(languageId),
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
		GranteeID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		WriterID:  cd.UserID,
	}); err != nil {
		return fmt.Errorf("update blog fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["PostURL"] = cd.AbsoluteURL(fmt.Sprintf("/blogs/blog/%d", row.Idblogs))
		}
	}

	return handlers.RedirectHandler(fmt.Sprintf("/blogs/blog/%d", row.Idblogs))
}

func BlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Edit Blog"
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	type Data struct {
		Languages          []*db.Language
		Blog               *db.GetBlogEntryForListerByIDRow
		SelectedLanguageId int
		Mode               string
	}

	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		Mode:               "Edit",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	data.Blog = cd.CurrentBlogLoaded()

	handlers.TemplateHandler(w, r, "blogEditPage.gohtml", data)
}
