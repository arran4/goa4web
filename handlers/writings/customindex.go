package writings

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/mux"
)

// WritingsCustomIndexItems returns the context-aware index items for writing pages.
func WritingsCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	items := []common.IndexItem{}

	path := "/writings"
	cd.RSSFeedURL = cd.GenerateFeedURL(path + "/rss")
	cd.AtomFeedURL = cd.GenerateFeedURL(path + "/atom")
	cd.PublicRSSFeedURL = path + "/rss"
	cd.PublicAtomFeedURL = path + "/atom"
	items = append(items,
		common.IndexItem{Name: "Atom Feed", Link: cd.AtomFeedURL},
		common.IndexItem{Name: "RSS Feed", Link: cd.RSSFeedURL},
	)

	items = append(items, common.IndexItem{
		Name: "Writings",
		Link: "/writings",
	})

	items = append(items, common.IndexItem{
		Name: "Writers",
		Link: "/writings/writers",
	})

	if cd.HasRole("content writer") {
		items = append(items, common.IndexItem{
			Name: "Write writings",
			Link: "/writings/add",
		})
	}

	if cd.IsAdmin() {
		items = append(items, common.IndexItem{
			Name: "Writings Admin",
			Link: "/admin/writings",
		})
	}

	vars := mux.Vars(r)
	writingIDStr := vars["writing"]

	if writingIDStr != "" {
		writing := cd.CurrentWritingLoaded()
		if writing != nil {
			// Mark as read
			items = append(items,
				common.IndexItem{
					Name: "Mark as read",
					Link: markWritingReadLink(writing.Idwriting, r.URL.RequestURI()),
				},
				common.IndexItem{
					Name: "Mark as read and go back",
					Link: markWritingReadLink(writing.Idwriting, fmt.Sprintf("/writings/category/%d", writing.WritingCategoryID)),
				},
			)

			// Edit link
			if cd.HasGrant("writing", "article", "edit", writing.Idwriting) {
				items = append(items, common.IndexItem{
					Name: "Edit Writing",
					Link: fmt.Sprintf("/writings/article/%d/edit", writing.Idwriting),
				})
			}
			// Admin Edit link
			if cd.IsAdmin() && cd.IsAdminMode() {
				items = append(items, common.IndexItem{
					Name: "Admin Edit Writing",
					Link: fmt.Sprintf("/admin/writings/article/%d", writing.Idwriting),
				})
			}
		}
	}

	return items
}

func markWritingReadLink(writingID int32, redirect string) string {
	link := fmt.Sprintf("/writings/article/%d/labels?task=%s", writingID, url.QueryEscape("Mark Thread Read"))
	if redirect != "" {
		link = fmt.Sprintf("%s&redirect=%s", link, url.QueryEscape(redirect))
	}
	return link
}
