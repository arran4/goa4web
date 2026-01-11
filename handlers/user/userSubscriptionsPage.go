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
	UserSubscriptionsPage.Handle(w, r, data)
}

const UserSubscriptionsPage handlers.Page = "user/subscriptions.gohtml"
