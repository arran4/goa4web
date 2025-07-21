package linker

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

func AdminAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Categories         []*db.LinkerCategory
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(config.AppRuntimeConfig.DefaultLanguage)),
	}

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "adminAddPage.gohtml", data)
}

type addTask struct{ tasks.TaskString }

var AddTask = &addTask{TaskString: TaskAdd}

// Compile-time interface conformance with context. When a link is submitted we
// alert subscribers of new content and notify administrators so they can review
// it for publication.
var _ tasks.Task = (*addTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*addTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*addTask)(nil)

func (addTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.CreateLinkerItem(r.Context(), db.CreateLinkerItemParams{
		UsersIdusers:     uid,
		LinkerCategoryID: int32(category),
		Title:            sql.NullString{Valid: true, String: title},
		Url:              sql.NullString{Valid: true, String: url},
		Description:      sql.NullString{Valid: true, String: description},
	}); err != nil {
		log.Printf("createLinkerItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)

}

func (addTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("linkerAddEmail")
}

func (addTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("linker_add")
	return &s
}

func (addTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLinkerAddEmail")
}

func (addTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLinkerAddEmail")
	return &v
}
