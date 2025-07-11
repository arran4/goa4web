// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries-password_resets.sql

package db

import (
	"context"
)

const createPasswordReset = `-- name: CreatePasswordReset :exec
INSERT INTO pending_passwords (user_id, passwd, passwd_algorithm, verification_code)
VALUES (?, ?, ?, ?)
`

type CreatePasswordResetParams struct {
	UserID           int32
	Passwd           string
	PasswdAlgorithm  string
	VerificationCode string
}

func (q *Queries) CreatePasswordReset(ctx context.Context, arg CreatePasswordResetParams) error {
	_, err := q.db.ExecContext(ctx, createPasswordReset,
		arg.UserID,
		arg.Passwd,
		arg.PasswdAlgorithm,
		arg.VerificationCode,
	)
	return err
}

const deletePasswordReset = `-- name: DeletePasswordReset :exec
DELETE FROM pending_passwords WHERE id = ?
`

func (q *Queries) DeletePasswordReset(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deletePasswordReset, id)
	return err
}

const getPasswordResetByCode = `-- name: GetPasswordResetByCode :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE verification_code = ? AND verified_at IS NULL
`

func (q *Queries) GetPasswordResetByCode(ctx context.Context, verificationCode string) (*PendingPassword, error) {
	row := q.db.QueryRowContext(ctx, getPasswordResetByCode, verificationCode)
	var i PendingPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Passwd,
		&i.PasswdAlgorithm,
		&i.VerificationCode,
		&i.CreatedAt,
		&i.VerifiedAt,
	)
	return &i, err
}

const getPasswordResetByUser = `-- name: GetPasswordResetByUser :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE user_id = ? AND verified_at IS NULL
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetPasswordResetByUser(ctx context.Context, userID int32) (*PendingPassword, error) {
	row := q.db.QueryRowContext(ctx, getPasswordResetByUser, userID)
	var i PendingPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Passwd,
		&i.PasswdAlgorithm,
		&i.VerificationCode,
		&i.CreatedAt,
		&i.VerifiedAt,
	)
	return &i, err
}

const markPasswordResetVerified = `-- name: MarkPasswordResetVerified :exec
UPDATE pending_passwords SET verified_at = NOW() WHERE id = ?
`

func (q *Queries) MarkPasswordResetVerified(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, markPasswordResetVerified, id)
	return err
}
