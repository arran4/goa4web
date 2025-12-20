package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// RolesQuerier implements role list queries for tests.
type RolesQuerier struct {
	*BaseQuerier
	Roles []*db.Role
}

// NewRolesQuerier returns a roles querier stub.
func NewRolesQuerier(t testing.TB) *RolesQuerier {
	t.Helper()
	return &RolesQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *RolesQuerier) AdminListRoles(ctx context.Context) ([]*db.Role, error) {
	return q.Roles, nil
}
