package user

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/subscriptions"
)

func userThreadSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Thread Subscriptions"

	dbThreads, err := cd.Queries().ListThreadSubscriptionsByUser(r.Context(), cd.UserID)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list subscriptions: %w", err))
		return
	}

	var dbSubs []*db.ListSubscriptionsByUserRow
	for _, sub := range dbThreads {
		dbSubs = append(dbSubs, &db.ListSubscriptionsByUserRow{
			ID:      sub.ID,
			Pattern: sub.Pattern,
			Method:  sub.Method,
		})
	}

	groups := subscriptions.GetUserSubscriptions(dbSubs)
	var threadGroups []*subscriptions.SubscriptionGroup
	for _, g := range groups {
		if g.Definition.Name == "Replies (Specific Thread)" {
			threadGroups = append(threadGroups, g)
		}
	}

	data := struct {
		Groups []*subscriptions.SubscriptionGroup
	}{
		Groups: threadGroups,
	}
	handlers.TemplateHandler(w, r, "user/subscriptions_threads.gohtml", data)
}
