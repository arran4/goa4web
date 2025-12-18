package news

import (
	"fmt"
	"net/http"
	"net/url"

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
	cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom" + suffix)
	cd.PublicRSSFeedURL = path + "/rss" + suffix
	cd.PublicAtomFeedURL = path + "/atom" + suffix

	items = append(items, common.IndexItem{
		Name:   "RSS Feed",
		Link:   "/news.rss",
		Folded: true,
	})

	userHasWriter := cd.HasGrant("news", "post", "post", 0)
	if userHasWriter {
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

		// Mark as Read
		redirect := r.URL.RequestURI()
		items = append(items, common.IndexItem{
			Name: "Mark as Read",
			Link: fmt.Sprintf("/news/news/%d/labels?task=Mark+Thread+Read&redirect=%s", post.Idsitenews, url.QueryEscape(redirect)),
		})
		items = append(items, common.IndexItem{
			Name: "Mark as Read & Return",
			Link: fmt.Sprintf("/news/news/%d/labels?task=Mark+Thread+Read&redirect=%s", post.Idsitenews, url.QueryEscape("/news")),
		})
	}
	return items
}

// Deprecated/Wrapper
func NewsCustomIndexItems(cd *common.CoreData, r *http.Request, post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow) []common.IndexItem {
	items := NewsGeneralIndexItems(cd, r)
	if post != nil {
		items = append(items, NewsPageSpecificItems(cd, r, post)...)
	}
	return items
}
