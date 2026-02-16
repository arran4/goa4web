package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveEmailTask updates email notification preferences.
type SaveEmailTask struct{ tasks.TaskString }

var saveEmailTask = &SaveEmailTask{TaskString: tasks.TaskString(TaskSaveAll)}

var _ tasks.Task = (*SaveEmailTask)(nil)

func (SaveEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return common.UserError{ErrorMessage: "forbidden"}
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	updates := r.PostFormValue("emailupdates") != ""
	auto := r.PostFormValue("autosubscribe") != ""

	if err := cd.SaveEmail(uid, updates, auto); err != nil {
		log.Printf("save email pref: %v", err)
		return fmt.Errorf("save email pref fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}
