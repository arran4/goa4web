package news

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"

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

func (EditTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationNewsEditEmail"), true
}

func (EditTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsEditEmail")
	return &v
}

func (EditTask) Page(w http.ResponseWriter, r *http.Request) { newsEditFormPage(w, r) }

func (EditTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["news"])
	if !cd.HasGrant("news", "post", "edit", int32(postId)) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}
	if err := cd.UpdateNewsPost(int32(postId), int32(languageId), cd.UserID, text); err != nil {
		return fmt.Errorf("update news post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func newsEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/news", http.StatusTemporaryRedirect)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/news?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if !cd.HasGrant("news", "post", "edit", post.Idsitenews) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	langs, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	cd.PageTitle = "Edit News"
	data := struct {
		Languages          []*db.Language
		Post               *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		SelectedLanguageId int
	}{
		Languages:          langs,
		Post:               post,
		SelectedLanguageId: int(post.LanguageID.Int32),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "newsEditPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
