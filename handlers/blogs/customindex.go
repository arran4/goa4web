package blogs

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
)

func BlogsGeneralIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	var items []common.IndexItem
	user := r.URL.Query().Get("user")

	if cd.FeedsEnabled {
		path := "/blogs"
		suffix := ""
		if user != "" {
			suffix = "?rss=" + url.QueryEscape(user)
		}
		cd.RSSFeedURL = cd.GenerateFeedURL(path + "/rss" + suffix)
		cd.RSSFeedTitle = "Blogs RSS Feed"
		cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom" + suffix)
		cd.AtomFeedTitle = "Blogs Atom Feed"
		cd.PublicRSSFeedURL = path + "/rss" + suffix
		cd.PublicAtomFeedURL = path + "/atom" + suffix

		items = append(items,
			common.IndexItem{Name: "Blogs Atom Feed", Link: cd.AtomFeedURL, Folded: true},
			common.IndexItem{Name: "Blogs RSS Feed", Link: cd.RSSFeedURL, Folded: true},
		)
	}

	if cd.IsAdmin() {
		items = append(items, common.IndexItem{
			Name: "Blogs Admin",
			Link: "/admin/blogs",
		})
	}
	userHasWriter := cd.HasRole("content writer")
	if userHasWriter {
		items = append(items, common.IndexItem{
			Name: "Write blog",
			Link: "/blogs/add",
		})
	}
	items = append(items, common.IndexItem{
		Name: "List bloggers",
		Link: "/blogs/bloggers",
	})
	return items
}

func BlogsPageSpecificItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	var items []common.IndexItem
	if blog, err := cd.BlogPost(); err == nil && blog != nil {
		if cd.CanEditBlog(blog.Idblogs, blog.UsersIdusers) {
			items = append(items, common.IndexItem{
				Name: "Edit Blog",
				Link: fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs),
			})
		}
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "Blog Admin",
				Link: fmt.Sprintf("/admin/blogs/blog/%d", blog.Idblogs),
			})
		}
	}
	return items
}

func BlogsMiddlewareIndex(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = append(cd.CustomIndexItems, BlogsGeneralIndexItems(cd, r)...)
}

// Deprecated: Use BlogsGeneralIndexItems and BlogsPageSpecificItems
func BlogsCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	items := BlogsGeneralIndexItems(cd, r)
	items = append(items, BlogsPageSpecificItems(cd, r)...)
	return items
}
