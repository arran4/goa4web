package blogs

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/mux"
)

// CustomBlogsIndex builds context-aware index items for blogs.
func CustomBlogsIndex(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = BlogsCustomIndexItems(cd, r)
}

// BlogsCustomIndexItems returns the context-aware index items for blog pages.
func BlogsCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	items := []common.IndexItem{}

	user := r.URL.Query().Get("user")
	vars := mux.Vars(r)
	blogIDStr := vars["blog"]

	// Feed links
	if cd.FeedsEnabled {
		path := "/blogs"
		suffix := ""
		if user != "" {
			suffix = "?rss=" + url.QueryEscape(user)
		}
		cd.RSSFeedURL = cd.GenerateFeedURL(path + "/rss" + suffix)
		cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom" + suffix)
		cd.PublicRSSFeedURL = path + "/rss" + suffix
		cd.PublicAtomFeedURL = path + "/atom" + suffix
		items = append(items,
			common.IndexItem{Name: "Atom Feed", Link: cd.AtomFeedURL},
			common.IndexItem{Name: "RSS Feed", Link: cd.RSSFeedURL},
		)
	}

	// List bloggers link (always present in old code)
	items = append(items, common.IndexItem{
		Name: "List bloggers",
		Link: "/blogs/bloggers",
	})

	// Admin link
	if cd.IsAdmin() {
		items = append(items, common.IndexItem{
			Name: "Blogs Admin",
			Link: "/admin/blogs",
		})
	}

	// Write blog link
	if cd.HasRole("content writer") {
		items = append(items, common.IndexItem{
			Name: "Write blog",
			Link: "/blogs/add",
		})
	}

	// Blog specific actions
	if blogIDStr != "" {
		blogEntry := cd.CurrentBlogLoaded()
		if blogEntry != nil {
			// Edit link
			if cd.CanEditBlog(blogEntry.Idblogs, blogEntry.UsersIdusers) {
				items = append(items, common.IndexItem{
					Name: "Edit Blog",
					Link: fmt.Sprintf("/blogs/blog/%d/edit", blogEntry.Idblogs),
				})
			}
			// Admin Edit link
			if cd.IsAdmin() && cd.IsAdminMode() {
				items = append(items, common.IndexItem{
					Name: "Admin Edit Blog",
					Link: fmt.Sprintf("/admin/blogs/blog/%d", blogEntry.Idblogs),
				})
			}
			// Comments link (if not on comments page? context implies we are on blog page)
			// The original template had [<a href="/blogs/blog/{{$blog.Idblogs}}/comments">{{$blog.Comments}} COMMENTS</a>]
			// Maybe add "View Comments" if we are not on the comments anchor?
			// The CustomIndex is for sidebar.
		}
	}

	return items
}
