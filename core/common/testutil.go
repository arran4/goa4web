package common

import (
	"context"
	"testing"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

// QuerierFake wraps an optional db.Querier and provides stubbed responses for common grant and topic lookups in tests.
type QuerierFake struct {
	db.Querier

	SystemCheckGrantErr       error
	SystemCheckGrantCalls     []db.SystemCheckGrantParams
	SystemCheckRoleGrantErr   error
	SystemCheckRoleGrantCalls []db.SystemCheckRoleGrantParams

	AdminListTopicsWithUserGrantsNoRolesRows  []*db.AdminListTopicsWithUserGrantsNoRolesRow
	AdminListTopicsWithUserGrantsNoRolesCalls []interface{}
}

// SystemCheckGrant records the call and returns a stubbed or delegated result.
func (q *QuerierFake) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	q.SystemCheckGrantCalls = append(q.SystemCheckGrantCalls, arg)
	if q.Querier != nil {
		return q.Querier.SystemCheckGrant(ctx, arg)
	}
	if q.SystemCheckGrantErr != nil {
		return 0, q.SystemCheckGrantErr
	}
	return 1, nil
}

// SystemCheckRoleGrant records the call and returns a stubbed or delegated result.
func (q *QuerierFake) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	q.SystemCheckRoleGrantCalls = append(q.SystemCheckRoleGrantCalls, arg)
	if q.Querier != nil {
		return q.Querier.SystemCheckRoleGrant(ctx, arg)
	}
	if q.SystemCheckRoleGrantErr != nil {
		return 0, q.SystemCheckRoleGrantErr
	}
	return 1, nil
}

// AdminListTopicsWithUserGrantsNoRoles records the call and returns stubbed topics or a delegated result.
func (q *QuerierFake) AdminListTopicsWithUserGrantsNoRoles(ctx context.Context, includeAdmin interface{}) ([]*db.AdminListTopicsWithUserGrantsNoRolesRow, error) {
	q.AdminListTopicsWithUserGrantsNoRolesCalls = append(q.AdminListTopicsWithUserGrantsNoRolesCalls, includeAdmin)
	if q.Querier != nil {
		return q.Querier.AdminListTopicsWithUserGrantsNoRoles(ctx, includeAdmin)
	}
	return q.AdminListTopicsWithUserGrantsNoRolesRows, nil
}

// NewTestCoreData returns a CoreData configured for tests using a QuerierFake.
// Use WithUserRoles or other CoreOption helpers to override defaults.
func NewTestCoreData(t *testing.T, q db.Querier) *CoreData {
	t.Helper()
	if q == nil {
		q = &QuerierFake{}
	}
	cache := &lazy.Value[[]*db.Role]{}
	return NewCoreData(context.Background(), q, config.NewRuntimeConfig(), WithRolesCache(cache))
}
