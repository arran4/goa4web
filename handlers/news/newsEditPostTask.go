package news

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type EditTask struct{ tasks.TaskString }

var editTask = &EditTask{TaskString: TaskEdit}

var _ tasks.Task = (*EditTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EditTask)(nil)

func (EditTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsEditEmail")
}

func (EditTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsEditEmail")
	return &v
}

func (EditTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["news"])

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("news", "post", "edit", int32(postId)) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}
	err = queries.UpdateNewsPostForWriter(r.Context(), db.UpdateNewsPostForWriterParams{
		PostID:      int32(postId),
		GrantPostID: sql.NullInt32{Int32: int32(postId), Valid: true},
		LanguageID:  int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
		GranteeID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		WriterID:  cd.UserID,
	})
	if err != nil {
		return fmt.Errorf("update news post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
