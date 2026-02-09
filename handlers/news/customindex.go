package news

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func NewsGeneralIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	var items []common.IndexItem

	// RSS Feed (Blind Apply)
	path := "/news"
	suffix := ""

	cd.RSSFeedURL = cd.GenerateFeedURL(path + "/rss" + suffix)
	cd.RSSFeedTitle = "News RSS Feed"
	cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom" + suffix)
	cd.AtomFeedTitle = "News Atom Feed"
	cd.PublicRSSFeedURL = path + "/rss" + suffix
	cd.PublicAtomFeedURL = path + "/atom" + suffix

	items = append(items, common.IndexItem{
		Name:   "News RSS Feed",
		Link:   cd.RSSFeedURL,
		Folded: true,
	})

	if CanPostNews(cd) {
		items = append(items, common.IndexItem{
			Name: "Add News",
			Link: "/news/post",
		})
	}

	vars := mux.Vars(r)
	newsId := vars["news"]
	if newsId != "" {
		items = append(items, common.IndexItem{
			Name: "Return to list",
			Link: fmt.Sprintf("/?offset=%d", 0),
		})
	}
	return items
}

func NewsPageSpecificItems(cd *common.CoreData, r *http.Request, post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow) []common.IndexItem {
	var items []common.IndexItem
	if post != nil {
		// Edit
		if cd.ShowEditNews(post.Idsitenews, post.UsersIdusers) {
			items = append(items, common.IndexItem{
				Name: "Edit News",
				Link: fmt.Sprintf("/news/news/%d/edit", post.Idsitenews),
			})
		}
		// Admin
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "News Admin",
				Link: fmt.Sprintf("/admin/news/article/%d", post.Idsitenews),
			})

			// Announcement
			ann := cd.NewsAnnouncement(post.Idsitenews)
			annText := "Promote to Announcement"
			if ann != nil && ann.Active {
				annText = "Manage Announcement"
			}
			items = append(items, common.IndexItem{
				Name: annText,
				Link: fmt.Sprintf("/admin/announcements?news_id=%d", post.Idsitenews),
			})
		}

		if hasNewsUnread(cd, post.Idsitenews, post.UsersIdusers) {
			redirect := r.URL.RequestURI()
			items = append(items, common.IndexItem{
				Name: "Mark as read",
				Link: fmt.Sprintf("/news/news/%d/labels?task=Mark+Thread+Read&redirect=%s", post.Idsitenews, url.QueryEscape(redirect)),
			})
			items = append(items, common.IndexItem{
				Name: "Mark as read and go back",
				Link: fmt.Sprintf("/news/news/%d/labels?task=Mark+Thread+Read&redirect=%s", post.Idsitenews, url.QueryEscape("/news")),
			})
		}
	}
	return items
}

func hasNewsUnread(cd *common.CoreData, postID int32, authorID int32) bool {
	if cd == nil || cd.UserID == 0 {
		return false
	}
	labels, err := cd.NewsPrivateLabels(postID, authorID)
	if err != nil {
		return false
	}
	for _, l := range labels {
		if l == "unread" || l == "new" {
			return true
		}
	}
	return false
}

// Deprecated/Wrapper
func NewsCustomIndexItems(cd *common.CoreData, r *http.Request, post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow) []common.IndexItem {
	items := NewsGeneralIndexItems(cd, r)
	if post != nil {
		items = append(items, NewsPageSpecificItems(cd, r, post)...)
	} else {
		vars := mux.Vars(r)
		if newsID := vars["news"]; newsID != "" {
			if nid, err := strconv.Atoi(newsID); err == nil {
				authorID := int32(0)
				// Fetch post if not provided to get author ID
				if p, err := cd.NewsPostByID(int32(nid)); err == nil && p != nil && p.Idusers.Valid {
					authorID = p.Idusers.Int32
				}
				if hasNewsUnread(cd, int32(nid), authorID) {
					redirect := r.URL.RequestURI()
					items = append(items, common.IndexItem{
						Name: "Mark as read",
						Link: fmt.Sprintf("/news/news/%d/labels?task=Mark+Thread+Read&redirect=%s", nid, url.QueryEscape(redirect)),
					})
				}
			}
		}
	}
	return items
}
