package user

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/subscriptions"
)

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Subscriptions"

	dbSubs, err := cd.Queries().ListSubscriptionsByUser(r.Context(), cd.UserID)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list subscriptions: %w", err))
		return
	}

	groups := subscriptions.GetUserSubscriptions(dbSubs)

	data := struct {
		Groups []*subscriptions.SubscriptionGroup
	}{
		Groups: groups,
	}
	handlers.TemplateHandler(w, r, "user/subscriptions.gohtml", data)
}

// UserSubscriptionUpdateTask updates the user's subscriptions.
// For now, this is a placeholder task to handle the form submission.
// Actual implementation will need to parse the form and call logic to add/remove DB rows.
// TODO: Implement the update logic.
