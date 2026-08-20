package forum

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/mux"
)

// CustomForumIndex builds context-aware index items for the public forum.
func CustomForumIndex(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = ForumCustomIndexItems(cd, r)
}

// ForumCustomIndexItems returns the context-aware index items for forum pages.
func ForumCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	base := forumBasePath(cd, r)
	section := "forum"
	if strings.HasPrefix(base, "/private") {
		section = "privateforum"
	}

	vars := mux.Vars(r)
	threadID := vars["thread"]
	topicID := vars["topic"]

	items := []common.IndexItem{}
	if cd.FeedsEnabled && topicID != "" && threadID == "" {
		cd.RSSFeedURL = fmt.Sprintf("%s/topic/%s.rss", base, topicID)
		cd.RSSFeedTitle = "Topic RSS Feed"
		cd.AtomFeedURL = fmt.Sprintf("%s/topic/%s.atom", base, topicID)
		cd.AtomFeedTitle = "Topic Atom Feed"
		items = append(items,
			common.IndexItem{Name: "Topic Atom Feed", Icon: "⚛️", Link: cd.AtomFeedURL, GroupID: "advanced"},
			common.IndexItem{Name: "Topic RSS Feed", Icon: "📡", Link: cd.RSSFeedURL, GroupID: "advanced"},
		)
	}

	if threadID != "" && topicID != "" {
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "[ADMIN]", Icon: "🛡️",
				Link: fmt.Sprintf("/admin/forum/topics/topic/%s", topicID),
			})
		}
		if hasThreadUnread(cd, threadID) {
			items = append(items,
				common.IndexItem{
					Name: "Mark as read", Icon: "✔️",
					Link: markThreadReadLink(base, threadID, r.URL.RequestURI()),
				},
				common.IndexItem{
					Name: "Mark as read and go back", Icon: "🔙", Link: markThreadReadLink(base, threadID, fmt.Sprintf("%s/topic/%s", base, topicID)),
				},
			)
		}
		if section != "privateforum" {
			items = append(items, common.IndexItem{
				Name: "Go to topic", Icon: "➡️", Link: fmt.Sprintf("%s/topic/%s", base, topicID),
			})
		}
		if tid, err := strconv.Atoi(topicID); err == nil && cd.HasGrant(section, "topic", "reply", int32(tid)) {
			items = append(items,
				common.IndexItem{
					Name: "Write Reply", Icon: "✍️",
					Link: fmt.Sprintf("%s/topic/%s/thread/%s#reply", base, topicID, threadID),
				},
			)
		}
		if tid, err := strconv.Atoi(topicID); err == nil && cd.HasGrant(section, "topic", "post", int32(tid)) {
			name := "New Thread"
			if base == "/private" {
				name = "Create a new private thread"
			}
			items = append(items,
				common.IndexItem{
					Name: name,
					Link: fmt.Sprintf("%s/topic/%s/thread", base, topicID),
				},
			)
		}
	}

	if threadID == "" && topicID != "" {
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "Admin Edit Topic", Icon: "⚙️",
				Link: fmt.Sprintf("/admin/forum/topics/topic/%s/edit", topicID),
			})
		}
		if tid, err := strconv.Atoi(topicID); err == nil && cd.HasGrant(section, "topic", "post", int32(tid)) {
			name := "New Thread"
			if base == "/private" {
				name = "Create a new private thread"
			}
			items = append(items,
				common.IndexItem{
					Name: name,
					Link: fmt.Sprintf("%s/topic/%s/thread", base, topicID),
				},
			)
		}
		if tid, err := strconv.Atoi(topicID); err == nil && cd.HasGrant(section, "topic", "label", int32(tid)) {
			items = append(items,
				common.IndexItem{
					Name: "Manage Labels", Icon: "🏷️", Link: fmt.Sprintf("%s/topic/%s/labels", base, topicID),
				},
			)
		}
		if cd.UserID != 0 {
			if tid, err := strconv.Atoi(topicID); err == nil {
				if subscribedToTopic(cd, int32(tid)) {
					items = append(items,
						common.IndexItem{
							Name:    "Unsubscribe From Topic",
							Link:    fmt.Sprintf("%s/topic/%s/unsubscribe", base, topicID),
							GroupID: "danger",
						},
					)
				} else {
					items = append(items,
						common.IndexItem{
							Name: "Subscribe To Topic", Icon: "🔔",
							Link: fmt.Sprintf("%s/topic/%s/subscribe", base, topicID),
						},
					)
				}
			}
		}
	}

	return items
}

func hasThreadUnread(cd *common.CoreData, threadID string) bool {
	if cd == nil || cd.UserID == 0 {
		return false
	}
	tid, err := strconv.Atoi(threadID)
	if err != nil {
		return false
	}
	var authorID int32
	if thread, err := cd.ForumThreadByID(int32(tid)); err == nil && thread != nil {
		if thread.Firstpostuserid.Valid {
			authorID = thread.Firstpostuserid.Int32
		}
	} else if err != nil {
		log.Printf("fetch thread %d: %v", tid, err)
	}

	labels, err := cd.ThreadPrivateLabels(int32(tid), authorID)
	if err != nil {
		log.Printf("thread private labels: %v", err)
		return false
	}
	for _, l := range labels {
		if l == "unread" || l == "new" {
			return true
		}
	}
	return false
}

func markThreadReadLink(base, threadID, redirect string) string {
	link := fmt.Sprintf("%s/thread/%s/labels?task=%s", base, threadID, url.QueryEscape(string(TaskMarkThreadRead)))
	if redirect != "" {
		link = fmt.Sprintf("%s&redirect=%s", link, url.QueryEscape(redirect))
	}
	return link
}

func forumBasePath(cd *common.CoreData, r *http.Request) string {
	if cd != nil && cd.ForumBasePath != "" {
		return cd.ForumBasePath
	}
	if r != nil && strings.HasPrefix(r.URL.Path, "/private") {
		return "/private"
	}
	return "/forum"
}
