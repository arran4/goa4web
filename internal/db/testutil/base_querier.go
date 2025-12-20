package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// BaseQuerier provides common permission helpers for test stubs.
type BaseQuerier struct {
	*UnimplementedQuerier
	GrantAllowed     bool
	RoleGrantAllowed bool
}

// NewBaseQuerier returns a base querier with default denied grants.
func NewBaseQuerier(t testing.TB) *BaseQuerier {
	t.Helper()
	return &BaseQuerier{
		UnimplementedQuerier: NewUnimplementedQuerier(t),
	}
}

// AllowGrants enables SystemCheckGrant responses.
func (q *BaseQuerier) AllowGrants() {
	q.GrantAllowed = true
}

// AllowRoleGrants enables SystemCheckRoleGrant responses.
func (q *BaseQuerier) AllowRoleGrants() {
	q.RoleGrantAllowed = true
}

// SystemCheckGrant implements db.Querier.
func (q *BaseQuerier) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if q.GrantAllowed {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

// SystemCheckRoleGrant implements db.Querier.
func (q *BaseQuerier) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	if q.RoleGrantAllowed {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}
