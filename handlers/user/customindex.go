package user

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for user pages. No items are added currently.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
	if cd.UserID != 0 {
		cd.CustomIndexItems = append(cd.CustomIndexItems,
			common.IndexItem{Name: "Language settings", Link: "/usr/lang"},
			common.IndexItem{Name: "Timezone settings", Link: "/usr/timezone"},
			common.IndexItem{Name: "Appearance settings", Link: "/usr/appearance"},
			common.IndexItem{Name: "Email and notification settings", Link: "/usr/email"},
			common.IndexItem{Name: "Your uploaded images", Link: "/usr/notifications/gallery"},
			common.IndexItem{Name: "Pagination settings", Link: "/usr/paging"},
			common.IndexItem{Name: "Subscriptions", Link: "/usr/subscriptions"},
			common.IndexItem{Name: "Public profile settings", Link: "/usr/profile"},
			common.IndexItem{Name: "API Keys", Link: "/usr/api-keys"},
		)
	}
}
