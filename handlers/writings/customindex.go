package writings

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
)

func WritingsGeneralIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	var items []common.IndexItem

	// RSS Feed
	path := "/writings"
	suffix := ""
	cd.RSSFeedURL = cd.GenerateFeedURL(path + "/rss" + suffix)
	cd.RSSFeedTitle = "Writings RSS Feed"
	cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom" + suffix)
	cd.AtomFeedTitle = "Writings Atom Feed"
	cd.PublicRSSFeedURL = path + "/rss" + suffix
	cd.PublicAtomFeedURL = path + "/atom" + suffix

	items = append(items,
		common.IndexItem{Name: "Writings Atom Feed", Link: "/writings/atom", Folded: true},
		common.IndexItem{Name: "Writings RSS Feed", Link: "/writings/rss", Folded: true},
	)

	if cd.IsAdmin() {
		items = append(items, common.IndexItem{
			Name: "Writings Admin",
			Link: "/admin/writings",
		})
	}
	userHasWriter := cd.HasContentWriterRole()
	if userHasWriter {
		items = append(items, common.IndexItem{
			Name: "Write writings",
			Link: "/writings/add",
		})
	}

	items = append(items, common.IndexItem{
		Name: "Writers",
		Link: "/writings/writers",
	})

	items = append(items, common.IndexItem{
		Name: "Return to list",
		Link: "/writings?offset=0",
	})
	return items
}

func WritingsPageSpecificItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	var items []common.IndexItem
	if writing, err := cd.Article(); err == nil && writing != nil {
		// Edit
		isAuthor := writing.UsersIdusers == cd.UserID
		canEdit := cd.HasContentWriterRole() && isAuthor
		if canEdit {
			items = append(items, common.IndexItem{
				Name: "Edit Writing",
				Link: fmt.Sprintf("/writings/article/%d/edit", writing.Idwriting),
			})
		}

		// Admin
		if cd.IsAdmin() && cd.IsAdminMode() {
			items = append(items, common.IndexItem{
				Name: "Writing Admin",
				Link: fmt.Sprintf("/admin/writings/article/%d", writing.Idwriting),
			})
		}

		// Mark as Read
		redirect := r.URL.RequestURI()
		items = append(items, common.IndexItem{
			Name: "Mark as Read",
			Link: fmt.Sprintf("/writings/article/%d/labels?task=Mark+Thread+Read&redirect=%s", writing.Idwriting, url.QueryEscape(redirect)),
		})
		items = append(items, common.IndexItem{
			Name: "Mark as Read & Return",
			Link: fmt.Sprintf("/writings/article/%d/labels?task=Mark+Thread+Read&redirect=%s", writing.Idwriting, url.QueryEscape(fmt.Sprintf("/writings/category/%d", writing.WritingCategoryID))),
		})
	}
	return items
}

// Deprecated/Wrapper
func WritingsCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	items := WritingsGeneralIndexItems(cd, r)
	items = append(items, WritingsPageSpecificItems(cd, r)...)
	return items
}
