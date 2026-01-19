package common

import (
	"database/sql"
	"errors"
	"log"

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

// AdminUserPendingPasswordResetCounts returns a map of user ID to count of pending password resets.
func (cd *CoreData) AdminUserPendingPasswordResetCounts() (map[int32]int64, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminCountPendingPasswordResetsByUser(cd.ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[int32]int64)
	for _, row := range rows {
		counts[row.UserID] = row.Count
	}
	return counts, nil
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

// AdminListPasswordResets lists password reset requests.
func (cd *CoreData) AdminListPasswordResets(status *string, limit, offset int32) ([]*db.AdminListPasswordResetsRow, int64, error) {
	if cd.queries == nil {
		return nil, 0, nil
	}
	var s sql.NullString
	if status != nil {
		s = sql.NullString{String: *status, Valid: true}
	}
	rows, err := cd.queries.AdminListPasswordResets(cd.ctx, db.AdminListPasswordResetsParams{
		Status: s,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	count, err := cd.queries.AdminCountPasswordResets(cd.ctx, db.AdminCountPasswordResetsParams{Status: s})
	if err != nil {
		return nil, 0, err
	}
	return rows, count, nil
}

// AdminApprovePasswordReset approves a password reset request.
func (cd *CoreData) AdminApprovePasswordReset(id int32) error {
	if cd.queries == nil {
		return errors.New("no queries")
	}
	reset, err := cd.queries.AdminGetPasswordResetByID(cd.ctx, id)
	if err != nil {
		return err
	}
	if reset.VerifiedAt.Valid {
		return errors.New("already verified")
	}
	// Mark as verified
	if err := cd.queries.SystemMarkPasswordResetVerified(cd.ctx, reset.ID); err != nil {
		return err
	}
	// Update user's password
	if err := cd.queries.InsertPassword(cd.ctx, db.InsertPasswordParams{
		UsersIdusers:    reset.UserID,
		Passwd:          reset.Passwd,
		PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true},
	}); err != nil {
		log.Printf("insert password: %v", err)
		return err
	}
	// Update request queue if exists
	_ = cd.queries.AdminUpdateRequestStatusByTableAndRow(cd.ctx, db.AdminUpdateRequestStatusByTableAndRowParams{
		Status:      "accepted",
		ChangeTable: "pending_passwords",
		ChangeRowID: id,
	})
	// Delete notification
	_ = cd.queries.AdminDeleteNotificationsByMessage(cd.ctx, sql.NullString{String: "Is attempting a password reset.", Valid: true})
	return nil
}

// AdminDenyPasswordReset denies (deletes) a password reset request.
func (cd *CoreData) AdminDenyPasswordReset(id int32) error {
	if cd.queries == nil {
		return errors.New("no queries")
	}
	// Update request queue if exists
	_ = cd.queries.AdminUpdateRequestStatusByTableAndRow(cd.ctx, db.AdminUpdateRequestStatusByTableAndRowParams{
		Status:      "rejected",
		ChangeTable: "pending_passwords",
		ChangeRowID: id,
	})
	// Delete notification
	_ = cd.queries.AdminDeleteNotificationsByMessage(cd.ctx, sql.NullString{String: "Is attempting a password reset.", Valid: true})

	return cd.queries.SystemDeletePasswordReset(cd.ctx, id)
}
