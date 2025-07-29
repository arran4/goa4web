package navigation

import (
	"sort"

	"github.com/arran4/goa4web/core/common"
)

// link represents a navigation item for either index or admin control center.
type link struct {
	section string
	name    string
	link    string
	weight  int
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

// RegisterAdminControlCenter registers a link for the admin control center menu in the given section.
func (r *Registry) RegisterAdminControlCenter(section, name, url string, weight int) {
	r.admin = append(r.admin, link{section: section, name: name, link: url, weight: weight})
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

// AdminSections returns admin navigation links grouped by section and sorted by weight.
func (r *Registry) AdminSections() []common.AdminSection {
	entries := make([]link, len(r.admin))
	copy(entries, r.admin)
	sort.Slice(entries, func(i, j int) bool { return entries[i].weight < entries[j].weight })

	secMap := map[string][]common.IndexItem{}
	order := []string{}
	for _, e := range entries {
		if _, ok := secMap[e.section]; !ok {
			secMap[e.section] = []common.IndexItem{}
			order = append(order, e.section)
		}
		secMap[e.section] = append(secMap[e.section], common.IndexItem{Name: e.name, Link: e.link})
	}

	sections := make([]common.AdminSection, 0, len(secMap))
	for _, sec := range order {
		sections = append(sections, common.AdminSection{Name: sec, Links: secMap[sec]})
	}
	return sections
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
func RegisterAdminControlCenter(section, name, url string, weight int) {
	defaultRegistry.RegisterAdminControlCenter(section, name, url, weight)
}

// IndexItems returns navigation items sorted by weight from the default registry.
func IndexItems() []common.IndexItem { return defaultRegistry.IndexItems() }

// AdminLinks returns admin navigation items sorted by weight from the default registry.
func AdminLinks() []common.IndexItem { return defaultRegistry.AdminLinks() }

// AdminSections returns admin navigation items grouped by section from the default registry.
func AdminSections() []common.AdminSection { return defaultRegistry.AdminSections() }
