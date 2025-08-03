package user

import (
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Subscriptions"
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	data := struct {
		Options []subscriptionOption
	}{
		Options: userSubscriptionOptions,
	}
	handlers.TemplateHandler(w, r, "subscriptions.gohtml", data)
}
