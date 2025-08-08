package common

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/arran4/goa4web/internal/db"
)

// ErrEmailAlreadyAssociated is returned when a user already has an email set.
var ErrEmailAlreadyAssociated = errors.New("email already associated")

// AssociateEmailParams carries the data required to request an email association.
type AssociateEmailParams struct {
	Username string
	Email    string
	Reason   string
}

// ErrPasswordResetRecentlyRequested indicates a password reset was requested too recently.
var ErrPasswordResetRecentlyRequested = errors.New("reset recently requested")

// UserCredentials fetches the stored password hash and algorithm for username.
func (cd *CoreData) UserCredentials(username string) (*db.SystemGetLoginRow, error) {
	if cd.queries == nil {
		return nil, fmt.Errorf("no queries available")
	}
	return cd.queries.SystemGetLogin(cd.ctx, sql.NullString{String: username, Valid: true})
}

// AssociateEmail creates an email association request for a user.
func (cd *CoreData) AssociateEmail(p AssociateEmailParams) (*db.SystemGetUserByUsernameRow, int64, error) {
	row, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: p.Username, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("user not found %w", err)
	}
	if row.Email != "" {
		return nil, 0, ErrEmailAlreadyAssociated
	}
	res, err := cd.queries.AdminInsertRequestQueue(cd.ctx, db.AdminInsertRequestQueueParams{
		UsersIdusers:   row.Idusers,
		ChangeTable:    "user_emails",
		ChangeField:    "email",
		ChangeRowID:    row.Idusers,
		ChangeValue:    sql.NullString{String: p.Email, Valid: true},
		ContactOptions: sql.NullString{String: p.Email, Valid: true},
	})
	if err != nil {
		log.Printf("insert admin request: %v", err)
		return nil, 0, fmt.Errorf("insert admin request %w", err)
	}
	id, _ := res.LastInsertId()
	_ = cd.queries.AdminInsertRequestComment(cd.ctx, db.AdminInsertRequestCommentParams{RequestID: int32(id), Comment: p.Reason})
	_ = cd.queries.InsertAdminUserComment(cd.ctx, db.InsertAdminUserCommentParams{UsersIdusers: row.Idusers, Comment: "email association requested"})
	return row, id, nil
}

// UserExists reports whether a user already exists with the supplied username or email address.
func (cd *CoreData) UserExists(username, email string) (bool, error) {
	if cd.queries == nil {
		return false, nil
	}
	if username != "" {
		if _, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true}); err == nil {
			return true, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("user by username: %w", err)
		}
	}
	if email != "" {
		if _, err := cd.queries.SystemGetUserByEmail(cd.ctx, email); err == nil {
			return true, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("user by email: %w", err)
		}
	}
	return false, nil
}

// CreateUserWithEmail inserts a user with the supplied username, email and password hash/algorithm, returning the new user ID.
func (cd *CoreData) CreateUserWithEmail(u, e, hash, alg string) (int32, error) {
	if cd.queries == nil {
		return 0, errors.New("no queries")
	}
	id, err := cd.queries.SystemInsertUser(cd.ctx, sql.NullString{String: u, Valid: u != ""})
	if err != nil {
		return 0, err
	}
	uid := int32(id)
	if err := cd.queries.InsertUserEmail(cd.ctx, db.InsertUserEmailParams{UserID: uid, Email: e, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{}}); err != nil {
		return 0, err
	}
	if err := cd.queries.InsertPassword(cd.ctx, db.InsertPasswordParams{UsersIdusers: uid, Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: alg != ""}}); err != nil {
		return 0, err
	}
	return uid, nil
}

// CreatePasswordReset creates a new password reset entry for the given email and returns the verification code.
func (cd *CoreData) CreatePasswordReset(email, hash, alg string) (string, error) {
	if cd.queries == nil {
		return "", nil
	}
	row, err := cd.queries.SystemGetUserByEmail(cd.ctx, email)
	if err != nil {
		return "", fmt.Errorf("user by email %w", err)
	}
	if reset, err := cd.queries.GetPasswordResetByUser(cd.ctx, db.GetPasswordResetByUserParams{
		UserID:    row.Idusers,
		CreatedAt: time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour),
	}); err == nil {
		if time.Since(reset.CreatedAt) < 24*time.Hour {
			return "", ErrPasswordResetRecentlyRequested
		}
		_ = cd.queries.SystemDeletePasswordReset(cd.ctx, reset.ID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get reset: %v", err)
		return "", fmt.Errorf("get reset %w", err)
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("rand %w", err)
	}
	code := hex.EncodeToString(buf[:])
	if err := cd.queries.CreatePasswordResetForUser(cd.ctx, db.CreatePasswordResetForUserParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		return "", fmt.Errorf("create reset %w", err)
	}
	return code, nil
}

// VerifyPasswordReset validates a reset code and activates the new password hash when successful.
func (cd *CoreData) VerifyPasswordReset(code string, newHash string) error {
	if cd.queries == nil {
		return errors.New("missing queries")
	}
	expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
	reset, err := cd.queries.GetPasswordResetByCode(cd.ctx, db.GetPasswordResetByCodeParams{VerificationCode: code, CreatedAt: expiry})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid code")
		}
		return fmt.Errorf("get reset %w", err)
	}
	if _, err := cd.queries.GetLoginRoleForUser(cd.ctx, reset.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("approval is pending")
		}
		return fmt.Errorf("user role %w", err)
	}
	if !verifyPassword(newHash, reset.Passwd, reset.PasswdAlgorithm) {
		return errors.New("invalid password")
	}
	if err := cd.queries.SystemMarkPasswordResetVerified(cd.ctx, reset.ID); err != nil {
		log.Printf("mark reset verified: %v", err)
	}
	if err := cd.queries.InsertPassword(cd.ctx, db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: reset.Passwd, PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true}}); err != nil {
		log.Printf("insert password: %v", err)
	}
	return nil
}

func verifyPassword(pw, storedHash, alg string) bool {
	parts := strings.Split(alg, ":")
	switch parts[0] {
	case "pbkdf2-sha256":
		if len(parts) != 3 {
			return false
		}
		iter, err := strconv.Atoi(parts[1])
		if err != nil {
			return false
		}
		salt, err := hex.DecodeString(parts[2])
		if err != nil {
			return false
		}
		hash := pbkdf2.Key([]byte(pw), salt, iter, 32, sha256.New)
		return storedHash == hex.EncodeToString(hash)
	case "md5", "":
		sum := md5.Sum([]byte(pw))
		return storedHash == hex.EncodeToString(sum[:])
	default:
		return false
	}
}
