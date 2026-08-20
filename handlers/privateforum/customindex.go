package privateforum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/gorilla/mux"
)

// CustomIndex injects private forum specific index items.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	vars := mux.Vars(r)
	topicID := vars["topic"]
	threadID := vars["thread"]
	items := []common.IndexItem{}

	if topicID == "" {
		items = []common.IndexItem{{
			Name: "Create New private topic", Icon: "🔒",
			Link: "/private/topic/new",
		}}
	} else {
		tid, err := strconv.Atoi(topicID)
		if err != nil {
			return
		}

		top, err := cd.ForumTopicByID(int32(tid))
		if err != nil || top == nil || top.Handler != "private" {
			return
		}

		if threadID != "" {
			items = append(items, common.IndexItem{
				Name: "Private threads",
				Icon: "📋",
				Link: fmt.Sprintf("/private/topic/%s", topicID),
			})
		}
		items = append(items, common.IndexItem{
			Name: "Private topics",
			Link: "/private",
		})

		if cd.HasGrant("privateforum", "topic", "edit", int32(tid)) {
			items = append(items, common.IndexItem{
				Name: "Edit Topic", Icon: "✏️", Link: fmt.Sprintf("/private/topic/%d/edit", tid),
			})
		}
	}

	unreadCountStr := ""
	tid := int32(0)
	if topicID != "" {
		if val, err := strconv.Atoi(topicID); err == nil {
			tid = int32(val)
		}
	}
	count, err := cd.UnreadPrivateThreadsCount(tid)
	if err == nil {
		if count > 0 {
			unreadCountStr = fmt.Sprintf(" (%d)", count)
		}

		link := "/private/unread"
		name := "All Unread"
		if tid > 0 {
			link = fmt.Sprintf("/private/topic/%d/unread", tid)
			name = "Unread in Topic"
		}

		if tid == 0 || count > 0 {
			items = append(items, common.IndexItem{
				Name: fmt.Sprintf("%s%s", name, unreadCountStr),
				Link: link,
			})
		}
	}

	items = append(items, forumhandlers.ForumCustomIndexItems(cd, r)...)
	cd.CustomIndexItems = items
}
