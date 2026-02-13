package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteEmailTask removes a user's email address.
type DeleteEmailTask struct{ tasks.TaskString }

var deleteEmailTask = &DeleteEmailTask{TaskString: tasks.TaskString(TaskDelete)}

var _ tasks.Task = (*DeleteEmailTask)(nil)

func (DeleteEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := cd.DeleteEmail(uid, int32(id)); err != nil {
		log.Printf("delete user email: %v", err)
		return fmt.Errorf("delete user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}
