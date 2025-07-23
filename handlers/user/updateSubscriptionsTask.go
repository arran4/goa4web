package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UpdateSubscriptionsTask saves user subscription preferences.
type UpdateSubscriptionsTask struct{ tasks.TaskString }

var updateSubscriptionsTask = &UpdateSubscriptionsTask{TaskString: tasks.TaskString(TaskUpdate)}

var _ tasks.Task = (*UpdateSubscriptionsTask)(nil)

func (UpdateSubscriptionsTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	existing, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		log.Printf("list subs: %v", err)
		return fmt.Errorf("list subscriptions fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	have := make(map[string]bool)
	for _, s := range existing {
		have[s.Pattern+"|"+s.Method] = true
	}
	methods := []string{"internal", "email"}
	for _, opt := range userSubscriptionOptions {
		for _, m := range methods {
			key := opt.Path + "_" + m
			want := r.PostFormValue(key) != ""
			hkey := opt.Pattern + "|" + m
			if want && !have[hkey] {
				if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m}); err != nil {
					log.Printf("insert sub: %v", err)
					return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			} else if !want && have[hkey] {
				if err := queries.DeleteSubscription(r.Context(), db.DeleteSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m}); err != nil {
					log.Printf("delete sub: %v", err)
					return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/subscriptions"}
}
