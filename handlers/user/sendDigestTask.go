package user

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type SendDigestTask struct{ tasks.TaskString }

var sendDigestNowTask = &SendDigestTask{TaskString: TaskSendDigestNow}
var sendDigestPreviewTask = &SendDigestTask{TaskString: TaskSendDigestPreview}

var _ tasks.Task = (*SendDigestTask)(nil)

func (t *SendDigestTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return handlers.SessionFetchFail{}
	}

	user, err := cd.Queries().SystemGetUserByID(r.Context(), uid)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	if !user.Email.Valid || user.Email.String == "" {
		return fmt.Errorf("user has no email")
	}

	pref, err := cd.Preference()
	if err != nil {
		return fmt.Errorf("fetch pref: %w", err)
	}
	markRead := false
	if pref != nil {
		markRead = pref.DailyDigestMarkRead
	}

	ep, ok := cd.EmailProvider().(email.Provider)
	if !ok {
		// If casting fails, it means the configured provider doesn't satisfy internal/email.Provider
		// This shouldn't happen if everything is wired correctly in workers.go
		return fmt.Errorf("email provider configuration error")
	}

	n := notifications.New(
		notifications.WithQueries(cd.Queries()),
		notifications.WithCustomQueries(cd.CustomQueries()),
		notifications.WithConfig(cd.Config),
		notifications.WithEmailProvider(ep),
	)

	preview := t.TaskString == TaskSendDigestPreview

	if err := n.SendDigestToUser(r.Context(), uid, user.Email.String, markRead, preview, notifications.DigestDaily); err != nil {
		return fmt.Errorf("send digest: %w", err)
	}

	return handlers.RefreshDirectHandler{TargetURL: "/usr/notifications"}
}
