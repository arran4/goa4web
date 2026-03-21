package privateforum

import (
	"net/http"

	forumhandlers "github.com/arran4/goa4web/handlers/forum"
)

// TopicPage displays a private topic with thread listings.
func TopicPage(w http.ResponseWriter, r *http.Request) {
	forumhandlers.TopicsPageWithBasePath(w, r, "/private")
}

// UnreadThreadsPage serves the private forum unread threads page.
func UnreadThreadsPage(w http.ResponseWriter, r *http.Request) {
	forumhandlers.UnreadThreadsPageWithBasePath(w, r, "/private")
}
