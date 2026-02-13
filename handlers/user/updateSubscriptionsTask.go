package user

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
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

		for _, m := range methods {
			if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: m}); err != nil {
				// We return the error here to inform the user if something went wrong (e.g. database error).
				// While this might surface "Duplicate entry" errors, it's better than failing silently.
				log.Printf("insert sub: %v", err)
				return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
				if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: method}); err != nil {
					log.Printf("insert sub: %v", err)
					return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			} else if !want && has {
				if err := queries.DeleteSubscriptionForSubscriber(r.Context(), db.DeleteSubscriptionForSubscriberParams{SubscriberID: uid, Pattern: pattern, Method: method}); err != nil {
					log.Printf("delete sub: %v", err)
					return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
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
					if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m}); err != nil {
						log.Printf("insert sub: %v", err)
						return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
					}
				} else if !want && have[hkey] {
					// Only delete if it matches one of the options in userSubscriptionOptions
					// This prevents deleting custom subscriptions added via the "Add" form
					if err := queries.DeleteSubscriptionForSubscriber(r.Context(), db.DeleteSubscriptionForSubscriberParams{SubscriberID: uid, Pattern: opt.Pattern, Method: m}); err != nil {
						log.Printf("delete sub: %v", err)
						return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
					}
				}
			}
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/subscriptions"}
}
