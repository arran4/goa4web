package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
)

func AdminAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Languages          []*db.Language
		SelectedLanguageId int
		Categories         []*db.LinkerCategory
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add Link"
	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusSeeOther)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "adminAddPage.gohtml", data)
}

type addTask struct{ tasks.TaskString }

var AdminAddTask = &addTask{TaskString: TaskAdd}

// Compile-time interface conformance with context. When a link is submitted we
// alert subscribers of new content and notify administrators so they can review
// it for publication.
var _ tasks.Task = (*addTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*addTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*addTask)(nil)

func (addTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return nil
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	allowed, err := UserCanCreateLink(r.Context(), queries, sql.NullInt32{Int32: int32(category), Valid: category != 0}, uid)
	if err != nil {
		return fmt.Errorf("UserCanCreateLink fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !allowed {
		return fmt.Errorf("UserCanCreateLink deny %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("forbidden")))
	}

	if err := queries.AdminCreateLinkerItem(r.Context(), db.AdminCreateLinkerItemParams{
		AuthorID:    uid,
		CategoryID:  sql.NullInt32{Int32: int32(category), Valid: category != 0},
		Title:       sql.NullString{Valid: true, String: title},
		Url:         sql.NullString{Valid: true, String: url},
		Description: sql.NullString{Valid: true, String: description},
		Listed:      sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Timezone:    sql.NullString{String: cd.Location().String(), Valid: true},
	}); err != nil {
		return fmt.Errorf("create linker item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func (addTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("linkerAddEmail"), true
}

func (addTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("linker_add")
	return &s
}

func (addTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationLinkerAddEmail"), true
}

func (addTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLinkerAddEmail")
	return &v
}
