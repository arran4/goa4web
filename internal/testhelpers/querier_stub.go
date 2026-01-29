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

	Permissions   []*db.GetPermissionsByUserIDRow
	PrivateLabels []*db.ListContentPrivateLabelsRow
	Subscriptions []*db.ListSubscriptionsByUserRow
}

// StubOption configures a test QuerierStub.
type StubOption func(*stubBuilder)

type stubBuilder struct {
	cfg            StubConfig
	grantAllowed   *bool
	grantError     error
	hasGrantReturn bool
}

// GrantKey returns the lookup key used to define grant expectations.
func GrantKey(section, item, action string) string {
	return section + "|" + item + "|" + action
}

// FromScenario adapts a reusable scenario into a stub option.
func FromScenario(scenario func(*stubBuilder)) StubOption {
	return func(builder *stubBuilder) {
		if scenario != nil {
			scenario(builder)
		}
	}
}

// ScenarioAdmin configures an admin permission set.
func ScenarioAdmin() func(*stubBuilder) {
	return func(builder *stubBuilder) {
		builder.cfg.Permissions = append(builder.cfg.Permissions, &db.GetPermissionsByUserIDRow{
			Name:    "administrator",
			IsAdmin: true,
		})
	}
}

// WithGrant marks a specific grant as allowed.
func WithGrant(section, item, action string) StubOption {
	return func(builder *stubBuilder) {
		if builder.cfg.Grants == nil {
			builder.cfg.Grants = map[string]bool{}
		}
		builder.cfg.Grants[GrantKey(section, item, action)] = true
	}
}

// WithGrants merges the provided grant map into the configuration.
func WithGrants(grants map[string]bool) StubOption {
	return func(builder *stubBuilder) {
		if len(grants) == 0 {
			return
		}
		if builder.cfg.Grants == nil {
			builder.cfg.Grants = map[string]bool{}
		}
		for key, allowed := range grants {
			builder.cfg.Grants[key] = allowed
		}
	}
}

// WithDefaultGrantAllowed sets the fallback grant decision when a specific key is not present.
func WithDefaultGrantAllowed(allowed bool) StubOption {
	return func(builder *stubBuilder) {
		builder.cfg.DefaultGrantAllowed = allowed
	}
}

// WithGrantResult forces SystemCheckGrant to return a fixed allowed/denied result.
func WithGrantResult(allowed bool) StubOption {
	return func(builder *stubBuilder) {
		builder.grantAllowed = &allowed
		builder.hasGrantReturn = true
		builder.grantError = nil
	}
}

// WithGrantError forces SystemCheckGrant to return the provided error.
func WithGrantError(err error) StubOption {
	return func(builder *stubBuilder) {
		builder.grantError = err
		builder.hasGrantReturn = true
		builder.grantAllowed = nil
	}
}

// WithPermissions sets the permissions returned by GetPermissionsByUserID.
func WithPermissions(permissions []*db.GetPermissionsByUserIDRow) StubOption {
	return func(builder *stubBuilder) {
		builder.cfg.Permissions = permissions
	}
}

// WithPrivateLabels sets the private labels returned by ListContentPrivateLabels.
func WithPrivateLabels(labels []*db.ListContentPrivateLabelsRow) StubOption {
	return func(builder *stubBuilder) {
		builder.cfg.PrivateLabels = labels
	}
}

// WithSubscriptions sets the subscriptions returned by ListSubscriptionsByUser.
func WithSubscriptions(subscriptions []*db.ListSubscriptionsByUserRow) StubOption {
	return func(builder *stubBuilder) {
		builder.cfg.Subscriptions = subscriptions
	}
}

// NewQuerierStub builds a db.QuerierStub configured with common grant and query responses.
func NewQuerierStub(options ...StubOption) *db.QuerierStub {
	builder := &stubBuilder{}
	for _, option := range options {
		if option != nil {
			option(builder)
		}
	}

	stub := &db.QuerierStub{
		GetPermissionsByUserIDReturns:   builder.cfg.Permissions,
		ListContentLabelStatusReturns:   []*db.ListContentLabelStatusRow{},
		ListContentPublicLabelsReturns:  []*db.ListContentPublicLabelsRow{},
		ListContentPrivateLabelsReturns: builder.cfg.PrivateLabels,
		ListSubscriptionsByUserReturns:  builder.cfg.Subscriptions,
		AddContentPrivateLabelIgnoreLabels: map[string]bool{
			"new":    true,
			"unread": true,
		},
	}

	if stub.GetPermissionsByUserIDReturns == nil {
		stub.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}
	}

	if stub.ListContentPrivateLabelsReturns == nil {
		stub.ListContentPrivateLabelsReturns = []*db.ListContentPrivateLabelsRow{}
	}

	if builder.hasGrantReturn {
		if builder.grantError != nil {
			stub.SystemCheckGrantErr = builder.grantError
		} else if builder.grantAllowed != nil && *builder.grantAllowed {
			stub.SystemCheckGrantReturns = 1
		}
		return stub
	}

	stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		item := ""
		if arg.Item.Valid {
			item = arg.Item.String
		}
		key := GrantKey(arg.Section, item, arg.Action)
		allowed, ok := builder.cfg.Grants[key]
		if !ok {
			allowed = builder.cfg.DefaultGrantAllowed
		}
		if allowed {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	return stub
}
