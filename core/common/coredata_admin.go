package common

import (
	"database/sql"
	"errors"

	"github.com/arran4/goa4web/internal/db"
)

// AdminListUsers returns all users with admin roles.
func (cd *CoreData) AdminListUsers() ([]*db.AdminListAllUsersRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminListAllUsers(cd.ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}

// AdminDashboardStats returns aggregate site statistics for the admin dashboard.
func (cd *CoreData) AdminDashboardStats() (*db.AdminGetDashboardStatsRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	stats, err := cd.queries.AdminGetDashboardStats(cd.ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return stats, nil
}

// AdminCommentsByUser lists all comments authored by the specified user.
func (cd *CoreData) AdminCommentsByUser(userID int32) ([]*db.AdminGetAllCommentsByUserRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminGetAllCommentsByUser(cd.ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}
