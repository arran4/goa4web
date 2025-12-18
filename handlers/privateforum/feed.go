package privateforum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/gorilla/mux"
)

func TopicRssPage(w http.ResponseWriter, r *http.Request) {
	TopicFeedHandler(w, r, "rss")
}

func TopicAtomPage(w http.ResponseWriter, r *http.Request) {
	TopicFeedHandler(w, r, "atom")
}

func TopicFeedHandler(w http.ResponseWriter, r *http.Request, feedType string) {
	// Private forum requires signature authentication for feeds
	// Session auth is technically possible but we prioritize/require signature for strictness
	// as requested: "always and only be a user signed rss feed"

	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	basePath := fmt.Sprintf("/private/topic/%d.%s", topicID, feedType)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Attempt signature verification
	if _, ok := vars["username"]; ok {
		user, err := handlers.VerifyFeedRequest(r, basePath)
		if err != nil {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Forbidden: %w", err))
			return
		}
		if user != nil {
			cd.UserID = user.Idusers
		}
	} else {
		// Strict mode: fail if not signed?
		// The requirement "always and only be a user signed rss feed" implies unsigned access should fail.
		// However, "GenerateFeedURL" only generates signed URL if logged in.
		// If I am browsing and click RSS, if I am logged in I get signed URL.
		// If I am NOT logged in, I get raw URL.
		// If raw URL is accessed, it should fail for private forum.
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	// Re-load user permissions/context with new UserID
	// CoreData methods usually check cd.UserID.

	topic, err := cd.ForumTopicByID(int32(topicID))
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	cd.PageTitle = fmt.Sprintf("Private Forum - %s Feed", topic.Title.String)
	rows, err := cd.ForumThreads(int32(topicID))
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	feed := forumhandlers.TopicFeed(r, topic.Title.String, topicID, rows, "/private")

	if feedType == "rss" {
		if err := feed.WriteRss(w); err != nil {
			handlers.RenderErrorPage(w, r, err)
		}
	} else {
		if err := feed.WriteAtom(w); err != nil {
			handlers.RenderErrorPage(w, r, err)
		}
	}
}
