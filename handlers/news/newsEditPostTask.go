package news

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
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

func (EditTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	raw := r.PostForm["author"]
	labels := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, l := range raw {
		if v := strings.TrimSpace(l); v != "" {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				labels = append(labels, v)
			}
		}
	}
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

	if err := cd.SetNewsAuthorLabels(int32(postId), labels); err != nil {
		return fmt.Errorf("set author labels fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
