package user

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/subscriptions"
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	// Check if this is an "Add Subscription" request
	if r.PostFormValue("task") == "Add" {
		definition := r.PostFormValue("definition")
		if definition == "" {
			return fmt.Errorf("missing definition %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("please select a subscription type")))
		}

		// Resolve parameters
		pattern := definition
		pattern = strings.ReplaceAll(pattern, "{topicid}", r.PostFormValue("topicid"))
		pattern = strings.ReplaceAll(pattern, "{threadid}", r.PostFormValue("threadid"))

		// Validate that no placeholders remain
		if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
			return fmt.Errorf("invalid parameters %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("missing required parameters")))
		}

		if def, _ := subscriptions.MatchDefinition(pattern); def != nil && def.IsAdminOnly && !cd.IsAdmin() {
			return fmt.Errorf("permission denied %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("cannot add admin subscription")))
		}

		methods := []string{}
		if r.PostFormValue("method_internal") != "" {
			methods = append(methods, "internal")
		}
		if r.PostFormValue("method_email") != "" {
			methods = append(methods, "email")
		}

		if len(methods) == 0 {
			return fmt.Errorf("no methods selected %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("please select at least one notification method")))
		}

		var inserts []db.InsertSubscriptionParams
		for _, m := range methods {
			inserts = append(inserts, db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: m})
		}

		if q, ok := queries.(*db.Queries); ok {
			if err := q.BatchInsertSubscriptions(r.Context(), inserts); err != nil {
				log.Printf("batch insert sub: %v", err)
				return fmt.Errorf("batch insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		} else {
			for _, p := range inserts {
				if err := queries.InsertSubscription(r.Context(), p); err != nil {
					log.Printf("insert sub: %v", err)
					return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
		}

		return handlers.RefreshDirectHandler{TargetURL: "/usr/subscriptions"}
	}

	existing, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		log.Printf("list subs: %v", err)
		return fmt.Errorf("list subscriptions fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	have := make(map[string]bool)
	for _, s := range existing {
		have[s.Pattern+"|"+s.Method] = true
	}

	var inserts []db.InsertSubscriptionParams
	var deletes []db.DeleteSubscriptionForSubscriberParams

	// New generic logic based on presented_subs
	presented := r.PostForm["presented_subs"]
	if len(presented) > 0 {
		wanted := make(map[string]bool)
		for _, s := range r.PostForm["subs"] {
			wanted[s] = true
		}

		for _, p := range presented {
			parts := strings.Split(p, "|")
			if len(parts) != 2 {
				continue
			}
			pattern, method := parts[0], parts[1]

			want := wanted[p]
			has := have[p]

			if def, _ := subscriptions.MatchDefinition(pattern); def != nil {
				if def.IsAdminOnly && !cd.IsAdmin() {
					continue
				}
				if def.Mandatory {
					want = true
				}
			}

			if want && !has {
				inserts = append(inserts, db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: method})
			} else if !want && has {
				deletes = append(deletes, db.DeleteSubscriptionForSubscriberParams{SubscriberID: uid, Pattern: pattern, Method: method})
			}
		}
	} else {
		methods := []string{"internal", "email"}
		for _, opt := range userSubscriptionOptions {
			for _, m := range methods {
				key := opt.Path + "_" + m
				want := r.PostFormValue(key) != ""
				hkey := opt.Pattern + "|" + m
				if want && !have[hkey] {
					inserts = append(inserts, db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m})
				} else if !want && have[hkey] {
					// Only delete if it matches one of the options in userSubscriptionOptions
					// This prevents deleting custom subscriptions added via the "Add" form
					deletes = append(deletes, db.DeleteSubscriptionForSubscriberParams{SubscriberID: uid, Pattern: opt.Pattern, Method: m})
				}
			}
		}
	}

	if q, ok := queries.(*db.Queries); ok {
		if err := q.BatchInsertSubscriptions(r.Context(), inserts); err != nil {
			log.Printf("batch insert sub: %v", err)
			return fmt.Errorf("batch insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if err := q.BatchDeleteSubscriptions(r.Context(), deletes); err != nil {
			log.Printf("batch delete sub: %v", err)
			return fmt.Errorf("batch delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	} else {
		for _, p := range inserts {
			if err := queries.InsertSubscription(r.Context(), p); err != nil {
				log.Printf("insert sub: %v", err)
				return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		for _, p := range deletes {
			if err := queries.DeleteSubscriptionForSubscriber(r.Context(), p); err != nil {
				log.Printf("delete sub: %v", err)
				return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/subscriptions"}
}
