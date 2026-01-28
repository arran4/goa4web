package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveDigestTask updates notification digest settings.
type SaveDigestTask struct{ tasks.TaskString }

var saveDigestTask = &SaveDigestTask{TaskString: TaskSaveDigest}

var _ tasks.Task = (*SaveDigestTask)(nil)

func (SaveDigestTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return common.UserError{ErrorMessage: "forbidden"}
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	hourStr := r.PostFormValue("daily_digest_hour")
	var hour *int
	if hourStr != "" && hourStr != "-1" {
		if h, err := strconv.Atoi(hourStr); err == nil && h >= 0 && h <= 23 {
			hour = &h
		}
	}

	markRead := r.PostFormValue("daily_digest_mark_read") == "on"

	if err := cd.SaveNotificationDigestPreferences(uid, hour, markRead); err != nil {
		log.Printf("save digest pref: %v", err)
		return fmt.Errorf("save digest pref fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/usr/notifications"}
}
