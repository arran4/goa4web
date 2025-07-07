package navigation

import (
	"sort"

	corecommon "github.com/arran4/goa4web/core/common"
)

// link represents a navigation item for either index or admin control center.
type link struct {
	name   string
	link   string
	weight int
}

var (
	indexRegistry []link
	adminRegistry []link
)

// RegisterIndexLink registers an entry for the site's index navigation.
func RegisterIndexLink(name, url string, weight int) {
	indexRegistry = append(indexRegistry, link{name: name, link: url, weight: weight})
}

// RegisterAdminControlCenter registers a link for the admin control center menu.
func RegisterAdminControlCenter(name, url string, weight int) {
	adminRegistry = append(adminRegistry, link{name: name, link: url, weight: weight})
}

// IndexItems returns navigation items sorted by weight.
func IndexItems() []corecommon.IndexItem {
	entries := make([]link, len(indexRegistry))
	copy(entries, indexRegistry)
	sort.Slice(entries, func(i, j int) bool { return entries[i].weight < entries[j].weight })
	items := make([]corecommon.IndexItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, corecommon.IndexItem{Name: e.name, Link: e.link})
	}
	return items
}

// AdminLinks returns admin navigation items sorted by weight.
func AdminLinks() []corecommon.IndexItem {
	entries := make([]link, len(adminRegistry))
	copy(entries, adminRegistry)
	sort.Slice(entries, func(i, j int) bool { return entries[i].weight < entries[j].weight })
	items := make([]corecommon.IndexItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, corecommon.IndexItem{Name: e.name, Link: e.link})
	}
	return items
}
