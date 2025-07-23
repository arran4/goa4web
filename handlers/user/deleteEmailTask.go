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

// DeleteEmailTask removes a user's email address.
type DeleteEmailTask struct{ tasks.TaskString }

var deleteEmailTask = &DeleteEmailTask{TaskString: tasks.TaskString(TaskDelete)}

var _ tasks.Task = (*DeleteEmailTask)(nil)

func (DeleteEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err == nil && ue.UserID == uid {
		if err := queries.DeleteUserEmail(r.Context(), int32(id)); err != nil {
			log.Printf("delete user email: %v", err)
			return fmt.Errorf("delete user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RedirectHandler("/usr/email")
}
