package user

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/subscriptions"
	"github.com/arran4/goa4web/internal/tasks"
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

	var filteredGroups []*subscriptions.SubscriptionGroup
	for _, g := range groups {
		if g.Definition.IsAdminOnly && !cd.IsAdmin() {
			continue
		}
		// Also ensure default instance for param-less definitions
		if len(g.Instances) == 0 && !strings.Contains(g.Definition.Pattern, "{") {
			// Create empty instance
			g.Instances = append(g.Instances, &subscriptions.SubscriptionInstance{
				Original:   "", // Will use definition pattern
				Methods:    []string{},
				Parameters: []subscriptions.Parameter{},
			})
		}
		filteredGroups = append(filteredGroups, g)
	}

	data := struct {
		Groups []*subscriptions.SubscriptionGroup
	}{
		Groups: filteredGroups,
	}
	UserSubscriptionsPage.Handle(w, r, data)
}

const UserSubscriptionsPage tasks.Template = "user/subscriptions.gohtml"
