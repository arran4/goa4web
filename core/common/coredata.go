package common

import "github.com/arran4/goa4web/internal/db"

// ContextValues represents context key names used across the application.
type ContextValues string

// IndexItem represents a navigation item linking to site sections.
type IndexItem struct {
	Name string
	Link string
}

type CoreData struct {
	IndexItems       []IndexItem
	CustomIndexItems []IndexItem
	UserID           int32
	// Username is the currently logged in user.
	Username      string
	SecurityLevel string
	Title         string
	AutoRefresh   bool
	FeedsEnabled  bool
	RSSFeedUrl    string
	AtomFeedUrl   string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	NotificationCount int32
	Announcement      *db.GetActiveAnnouncementWithNewsRow
}

var rolePriority = map[string]int{
	"reader":        1,
	"writer":        2,
	"moderator":     3,
	"administrator": 4,
}

func (cd *CoreData) HasRole(role string) bool {
	return rolePriority[cd.SecurityLevel] >= rolePriority[role]
}

// ContainsItem returns true if items includes an entry with the given name.
func ContainsItem(items []IndexItem, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}
