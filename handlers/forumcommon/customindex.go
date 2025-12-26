package forumcommon

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

// ForumCustomIndexItems returns the context-aware index items for forum pages.
func ForumCustomIndexItems(cd *common.CoreData, r *http.Request, section string) []common.IndexItem {
	base := ForumBasePath(cd, r)
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
			common.IndexItem{Name: "Topic Atom Feed", Link: cd.AtomFeedURL, Folded: true},
			common.IndexItem{Name: "Topic RSS Feed", Link: cd.RSSFeedURL, Folded: true},
		)
	}

	if threadID != "" && topicID != "" {
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "[ADMIN]",
				Link: fmt.Sprintf("/admin/forum/topics/topic/%s", topicID),
			})
		}
		if HasThreadUnread(cd, threadID) {
			items = append(items,
				common.IndexItem{
					Name: "Mark as read",
					Link: MarkThreadReadLink(base, threadID, r.URL.RequestURI()),
				},
				common.IndexItem{
					Name: "Mark as read and go back",
					Link: MarkThreadReadLink(base, threadID, fmt.Sprintf("%s/topic/%s", base, topicID)),
				},
			)
		}
		items = append(items, common.IndexItem{
			Name: "Go to topic",
			Link: fmt.Sprintf("%s/topic/%s", base, topicID),
		})
		if tid, err := strconv.Atoi(topicID); err == nil && cd.HasGrant(section, "topic", "reply", int32(tid)) {
			items = append(items,
				common.IndexItem{
					Name: "Write Reply",
					Link: fmt.Sprintf("%s/topic/%s/thread/%s/reply", base, topicID, threadID),
				},
			)
		}
	}

	if threadID == "" && topicID != "" {
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "Admin Edit Topic",
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
		if cd.UserID != 0 {
			if tid, err := strconv.Atoi(topicID); err == nil {
				if SubscribedToTopic(cd, int32(tid)) {
					items = append(items,
						common.IndexItem{
							Name:   "Unsubscribe From Topic",
							Link:   fmt.Sprintf("%s/topic/%s/unsubscribe", base, topicID),
							Folded: true,
						},
					)
				} else {
					items = append(items,
						common.IndexItem{
							Name: "Subscribe To Topic",
							Link: fmt.Sprintf("%s/topic/%s/subscribe", base, topicID),
						},
					)
				}
			}
		}
	}

	return items
}

func HasThreadUnread(cd *common.CoreData, threadID string) bool {
	if cd == nil || cd.UserID == 0 {
		return false
	}
	tid, err := strconv.Atoi(threadID)
	if err != nil {
		return false
	}
	labels, err := cd.ThreadPrivateLabels(int32(tid))
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

func MarkThreadReadLink(base, threadID, redirect string) string {
	link := fmt.Sprintf("%s/thread/%s/labels?task=%s", base, threadID, url.QueryEscape(string(TaskMarkThreadRead)))
	if redirect != "" {
		link = fmt.Sprintf("%s&redirect=%s", link, url.QueryEscape(redirect))
	}
	return link
}

func ForumBasePath(cd *common.CoreData, r *http.Request) string {
	if cd != nil && cd.ForumBasePath != "" {
		return cd.ForumBasePath
	}
	if r != nil && strings.HasPrefix(r.URL.Path, "/private") {
		return "/private"
	}
	return "/forum"
}
