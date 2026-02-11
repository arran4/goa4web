package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

type fakeForgotPasswordQueries struct {
	db.Querier
	loginRow       *db.SystemGetLoginRow
	verifiedEmails []string
	roleErr        error
	userByEmail    *db.SystemGetUserByEmailRow
	pending        *db.PendingPassword
	deleteID       int32
	createdResets  []db.CreatePasswordResetForUserParams
}

func (f *fakeForgotPasswordQueries) SystemGetLogin(_ context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
	if f.loginRow != nil && username.String == f.loginRow.Username.String {
		return f.loginRow, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeForgotPasswordQueries) SystemListVerifiedEmailsByUserID(_ context.Context, userID int32) ([]*db.UserEmail, error) {
	rows := make([]*db.UserEmail, 0, len(f.verifiedEmails))
	for i, email := range f.verifiedEmails {
		rows = append(rows, &db.UserEmail{
			ID:                    int32(i + 1),
			UserID:                userID,
			Email:                 email,
			VerifiedAt:            sql.NullTime{Time: time.Now(), Valid: true},
			NotificationPriority:  int32(i),
			LastVerificationCode:  sql.NullString{},
			VerificationExpiresAt: sql.NullTime{},
		})
	}
	return rows, nil
}

func (f *fakeForgotPasswordQueries) GetLoginRoleForUser(_ context.Context, _ int32) (int32, error) {
	if f.roleErr != nil {
		return 0, f.roleErr
	}
	return 1, nil
}

func (f *fakeForgotPasswordQueries) SystemGetUserByEmail(_ context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
	if f.userByEmail != nil && f.userByEmail.Email == email {
		return f.userByEmail, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeForgotPasswordQueries) GetPasswordResetByUser(_ context.Context, _ db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
	if f.pending == nil {
		return nil, sql.ErrNoRows
	}
	return f.pending, nil
}

func (f *fakeForgotPasswordQueries) SystemDeletePasswordReset(_ context.Context, id int32) error {
	f.deleteID = id
	return nil
}

func (f *fakeForgotPasswordQueries) CreatePasswordResetForUser(_ context.Context, arg db.CreatePasswordResetForUserParams) error {
	f.createdResets = append(f.createdResets, arg)
	return nil
}

func (f *fakeForgotPasswordQueries) AdminInsertRequestQueue(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
	return nil, nil
}
