package user

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/subscriptions"
)

func userSubscriptionAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add Subscription"

	data := struct {
		Definitions []subscriptions.Definition
	}{
		Definitions: subscriptions.Definitions,
	}
	handlers.TemplateHandler(w, r, "user/subscription_add.gohtml", data)
}
