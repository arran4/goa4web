package privateforum

import (
	"net/http"

	forumhandlers "github.com/arran4/goa4web/handlers/forum"
)

// ThreadPage displays a thread within a private topic.
func ThreadPage(w http.ResponseWriter, r *http.Request) {
	forumhandlers.ThreadPageWithBasePath(w, r, "/private")
}
