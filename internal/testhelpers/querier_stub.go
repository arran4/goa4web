package testhelpers

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// StubConfig describes the preconfigured responses for a QuerierStub used in tests.
type StubConfig struct {
	Grants map[string]bool
	// DefaultGrantAllowed controls the fallback grant decision when a specific key is not present.
	DefaultGrantAllowed bool

	PrivateLabels []*db.ListContentPrivateLabelsRow
	Subscriptions []*db.ListSubscriptionsByUserRow
}

// GrantKey returns the lookup key used to define grant expectations.
func GrantKey(section, item, action string) string {
	return section + "|" + item + "|" + action
}

// NewQuerierStub builds a db.QuerierStub configured with common grant and query responses.
func NewQuerierStub(cfg StubConfig) *db.QuerierStub {
	stub := &db.QuerierStub{
		ListContentPrivateLabelsReturns: cfg.PrivateLabels,
		ListSubscriptionsByUserReturns:  cfg.Subscriptions,
	}

	stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		item := ""
		if arg.Item.Valid {
			item = arg.Item.String
		}
		key := GrantKey(arg.Section, item, arg.Action)
		allowed, ok := cfg.Grants[key]
		if !ok {
			allowed = cfg.DefaultGrantAllowed
		}
		if allowed {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	return stub
}
