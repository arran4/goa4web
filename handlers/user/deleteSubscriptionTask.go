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
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteSubscriptionTask removes a subscription entry.
type DeleteTask struct{ tasks.TaskString }

var deleteTask = &DeleteTask{TaskString: tasks.TaskString(TaskDelete)}

var _ tasks.Task = (*DeleteTask)(nil)

func (DeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	idStr := r.PostFormValue("id")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if idStr == "" {
		return handlers.RedirectHandler("/usr/subscriptions?error=missing id")
	}
	id, _ := strconv.Atoi(idStr)
	if err := queries.DeleteSubscriptionByID(r.Context(), db.DeleteSubscriptionByIDParams{UsersIdusers: uid, ID: int32(id)}); err != nil {
		log.Printf("delete sub: %v", err)
		return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler("/usr/subscriptions")
}
