package common

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

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
	Title            string
	AutoRefresh      bool
	FeedsEnabled     bool
	RSSFeedUrl       string
	AtomFeedUrl      string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	NotificationCount int32
	a4codeMapper      func(tag, val string) string

	session *sessions.Session

	ctx     context.Context
	queries *db.Queries

	user         lazyValue[*db.User]
	perms        lazyValue[[]*db.GetUserRolesRow]
	pref         lazyValue[*db.Preference]
	langs        lazyValue[[]*db.UserLanguage]
	roles        lazyValue[[]string]
	announcement lazyValue[*db.GetActiveAnnouncementWithNewsRow]

	event *eventbus.Event
}

// SetRole preloads the current role value.
func (cd *CoreData) SetRoles(r []string) { cd.roles.set(r) }

// CoreOption configures a new CoreData instance.
type CoreOption func(*CoreData)

// WithImageURLMapper sets the a4code image mapper option.
func WithImageURLMapper(fn func(tag, val string) string) CoreOption {
	return func(cd *CoreData) { cd.a4codeMapper = fn }
}

// WithSession stores the gorilla session on the CoreData object.
func WithSession(s *sessions.Session) CoreOption {
	return func(cd *CoreData) { cd.session = s }
}

// WithEvent links an event to the CoreData object.
func WithEvent(evt *eventbus.Event) CoreOption { return func(cd *CoreData) { cd.event = evt } }

// NewCoreData creates a CoreData with context and queries applied.
func NewCoreData(ctx context.Context, q *db.Queries, opts ...CoreOption) *CoreData {
	cd := &CoreData{ctx: ctx, queries: q}
	for _, o := range opts {
		o(cd)
	}
	return cd
}

// ImageURLMapper maps image references like "image:" or "cache:" to full URLs.
func (cd *CoreData) ImageURLMapper(tag, val string) string {
	if cd.a4codeMapper != nil {
		return cd.a4codeMapper(tag, val)
	}
	return val
}

var RolePriority = map[string]int{
	"reader":        1,
	"writer":        2,
	"moderator":     3,
	"administrator": 4,
}

func (cd *CoreData) HasRole(role string) bool {
	return RolePriority[cd.Role()] >= RolePriority[role]
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

// Role returns the user role loaded lazily.
func (cd *CoreData) Roles() []string {
	roles, _ := cd.roles.load(func() ([]string, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return []string{"reader"}, nil
		}
		perms, err := cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
		if err != nil {
			return []string{"reader"}, nil
		}
		var rs []string
		for _, p := range perms {
			if p.Role != "" {
				rs = append(rs, p.Role)
			}
		}
		if len(rs) == 0 {
			rs = []string{"reader"}
		}
		return rs, nil
	})
	return roles
}

func (cd *CoreData) Role() string {
	roles := cd.Roles()
	best := "reader"
	for _, r := range roles {
		if RolePriority[r] > RolePriority[best] {
			best = r
		}
	}
	return best
}

// SetSession stores s on cd for later retrieval.
func (cd *CoreData) SetSession(s *sessions.Session) { cd.session = s }

// Session returns the request session if available.
func (cd *CoreData) Session() *sessions.Session { return cd.session }

// SetEvent stores evt on cd for handler access.
func (cd *CoreData) SetEvent(evt *eventbus.Event) { cd.event = evt }

// Event returns the event associated with the request, if any.
func (cd *CoreData) Event() *eventbus.Event { return cd.event }

// CurrentUser returns the logged in user's record loaded on demand.
func (cd *CoreData) CurrentUser() (*db.User, error) {
	return cd.user.load(func() (*db.User, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetUserById(cd.ctx, cd.UserID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			return nil, nil
		}
		return &db.User{Idusers: row.Idusers, Username: row.Username}, nil
	})
}

// Permissions returns the user's permissions loaded on demand.
func (cd *CoreData) Permissions() ([]*db.GetUserRolesRow, error) {
	return cd.perms.load(func() ([]*db.GetUserRolesRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
	})
}

// Preference returns the user's preferences loaded on demand.
func (cd *CoreData) Preference() (*db.Preference, error) {
	return cd.pref.load(func() (*db.Preference, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPreferenceByUserID(cd.ctx, cd.UserID)
	})
}

// Languages returns the user's language selections loaded on demand.
func (cd *CoreData) Languages() ([]*db.UserLanguage, error) {
	return cd.langs.load(func() ([]*db.UserLanguage, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetUserLanguages(cd.ctx, cd.UserID)
	})
}

// Announcement returns the active announcement row loaded lazily.
func (cd *CoreData) Announcement() *db.GetActiveAnnouncementWithNewsRow {
	ann, _ := cd.announcement.load(func() (*db.GetActiveAnnouncementWithNewsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetActiveAnnouncementWithNews(cd.ctx)
		if err != nil {
			return nil, err
		}
		return row, nil
	})
	return ann
}
