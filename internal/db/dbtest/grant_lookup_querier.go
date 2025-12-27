package dbtest

import (
	"context"
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// GrantLookupQuerier fakes grant checks and user lookups for permission tests.
type GrantLookupQuerier struct {
	db.Querier

	// GrantResults configures the response returned by SystemCheckGrant.
	GrantResults []error

	CheckUserHasGrantReturn bool
	CheckUserHasGrantErr    error

	SystemGetUserByUsernameRow   *db.SystemGetUserByUsernameRow
	SystemGetUserByUsernameErr   error
	SystemGetUserByUsernameCalls []sql.NullString
	SystemCheckGrantCalls        []db.SystemCheckGrantParams
	CheckUserHasGrantCalls       []db.CheckUserHasGrantParams
	GetPermissionsByUserIDReturn []*db.GetPermissionsByUserIDRow
	GetPermissionsByUserIDErr    error
	GetPermissionsByUserIDCalls  []int32
	RoleGrantResults             []error
	SystemCheckRoleGrantCalls    []db.SystemCheckRoleGrantParams
}

// SystemCheckGrant records grant lookups and returns the next configured result.
func (q *GrantLookupQuerier) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	q.SystemCheckGrantCalls = append(q.SystemCheckGrantCalls, arg)
	if len(q.GrantResults) == 0 || q.GrantResults[0] == nil {
		if len(q.GrantResults) > 0 {
			q.GrantResults = q.GrantResults[1:]
		}
		return 1, nil
	}
	err := q.GrantResults[0]
	q.GrantResults = q.GrantResults[1:]
	return 0, err
}

// CheckUserHasGrant records grant rule queries and returns the configured values.
func (q *GrantLookupQuerier) CheckUserHasGrant(ctx context.Context, arg db.CheckUserHasGrantParams) (bool, error) {
	q.CheckUserHasGrantCalls = append(q.CheckUserHasGrantCalls, arg)
	return q.CheckUserHasGrantReturn, q.CheckUserHasGrantErr
}

// SystemCheckRoleGrant records role grant lookups and returns the next configured result.
func (q *GrantLookupQuerier) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	q.SystemCheckRoleGrantCalls = append(q.SystemCheckRoleGrantCalls, arg)
	if len(q.RoleGrantResults) == 0 || q.RoleGrantResults[0] == nil {
		if len(q.RoleGrantResults) > 0 {
			q.RoleGrantResults = q.RoleGrantResults[1:]
		}
		return 1, nil
	}
	err := q.RoleGrantResults[0]
	q.RoleGrantResults = q.RoleGrantResults[1:]
	return 0, err
}

// GetPermissionsByUserID records permission lookups and returns the configured result.
func (q *GrantLookupQuerier) GetPermissionsByUserID(ctx context.Context, userID int32) ([]*db.GetPermissionsByUserIDRow, error) {
	q.GetPermissionsByUserIDCalls = append(q.GetPermissionsByUserIDCalls, userID)
	if q.GetPermissionsByUserIDErr != nil {
		return nil, q.GetPermissionsByUserIDErr
	}
	return q.GetPermissionsByUserIDReturn, nil
}

// SystemGetUserByUsername records username lookups and returns the configured result.
func (q *GrantLookupQuerier) SystemGetUserByUsername(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
	q.SystemGetUserByUsernameCalls = append(q.SystemGetUserByUsernameCalls, username)
	if q.SystemGetUserByUsernameErr != nil {
		return nil, q.SystemGetUserByUsernameErr
	}
	return q.SystemGetUserByUsernameRow, nil
}
