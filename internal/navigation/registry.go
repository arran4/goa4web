package navigation

import (
	"sort"

	"github.com/arran4/goa4web/core/common"
)

// link represents a navigation item for either index or admin control center.
type link struct {
	name   string
	link   string
	weight int
}

// Registry stores navigation entries for the public index and admin pages.
type Registry struct {
	index []link
	admin []link
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// RegisterIndexLink registers an entry for the site's index navigation.
func (r *Registry) RegisterIndexLink(name, url string, weight int) {
	r.index = append(r.index, link{name: name, link: url, weight: weight})
}

// RegisterAdminControlCenter registers a link for the admin control center menu.
func (r *Registry) RegisterAdminControlCenter(name, url string, weight int) {
	r.admin = append(r.admin, link{name: name, link: url, weight: weight})
}

// IndexItems returns navigation items sorted by weight.
func (r *Registry) IndexItems() []common.IndexItem {
	entries := make([]link, len(r.index))
	copy(entries, r.index)
	sort.Slice(entries, func(i, j int) bool { return entries[i].weight < entries[j].weight })
	items := make([]common.IndexItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, common.IndexItem{Name: e.name, Link: e.link})
	}
	return items
}

// AdminLinks returns admin navigation items sorted by weight.
func (r *Registry) AdminLinks() []common.IndexItem {
	entries := make([]link, len(r.admin))
	copy(entries, r.admin)
	sort.Slice(entries, func(i, j int) bool { return entries[i].weight < entries[j].weight })
	items := make([]common.IndexItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, common.IndexItem{Name: e.name, Link: e.link})
	}
	return items
}

var defaultRegistry = NewRegistry()

// SetDefaultRegistry sets the package level registry used by the helper functions.
func SetDefaultRegistry(r *Registry) {
	if r != nil {
		defaultRegistry = r
	}
}

// RegisterIndexLink registers an entry for the site's index navigation using the default registry.
func RegisterIndexLink(name, url string, weight int) {
	defaultRegistry.RegisterIndexLink(name, url, weight)
}

// RegisterAdminControlCenter registers a link for the admin control center menu using the default registry.
func RegisterAdminControlCenter(name, url string, weight int) {
	defaultRegistry.RegisterAdminControlCenter(name, url, weight)
}

// IndexItems returns navigation items sorted by weight from the default registry.
func IndexItems() []common.IndexItem { return defaultRegistry.IndexItems() }

// AdminLinks returns admin navigation items sorted by weight from the default registry.
func AdminLinks() []common.IndexItem { return defaultRegistry.AdminLinks() }
