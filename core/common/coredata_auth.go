package common

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
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
		return nil, fmt.Errorf("CoreData.UserCredentials: no queries available")
	}
	return cd.queries.SystemGetLogin(cd.ctx, sql.NullString{String: username, Valid: true})
}

// VerifiedEmailsForUser returns all verified email addresses ordered by notification priority.
func (cd *CoreData) VerifiedEmailsForUser(userID int32) ([]string, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.SystemListVerifiedEmailsByUserID(cd.ctx, userID)
	if err != nil {
		return nil, err
	}
	emails := make([]string, 0, len(rows))
	for _, row := range rows {
		emails = append(emails, row.Email)
	}
	return emails, nil
}

// AssociateEmail creates an email association request for a user.
func (cd *CoreData) AssociateEmail(p AssociateEmailParams) (*db.SystemGetUserByUsernameRow, int64, error) {
	row, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: p.Username, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("CoreData.AssociateEmail: user not found %w", err)
	}
	verifiedEmails, err := cd.VerifiedEmailsForUser(row.Idusers)
	if err != nil {
		return nil, 0, fmt.Errorf("CoreData.AssociateEmail: list verified emails %w", err)
	}
	if len(verifiedEmails) > 0 {
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
		return nil, 0, fmt.Errorf("CoreData.AssociateEmail: insert admin request %w", err)
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
			return false, fmt.Errorf("CoreData.UserExists: user by username: %w", err)
		}
	}
	if email != "" {
		if _, err := cd.queries.SystemGetUserByEmail(cd.ctx, email); err == nil {
			return true, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("CoreData.UserExists: user by email: %w", err)
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
		return "", fmt.Errorf("CoreData.CreatePasswordReset: user by email %w", err)
	}
	return cd.CreatePasswordResetForUser(row.Idusers, hash, alg)
}

// CreatePasswordResetForUser creates a new password reset entry for the given user ID and returns the verification code.
func (cd *CoreData) CreatePasswordResetForUser(userID int32, hash, alg string) (string, error) {
	if cd.queries == nil {
		return "", nil
	}
	// Check for existing recent request (only for standard user requests, admin generated ones are handled via signed URLs mostly, but we still create a record)
	// If hash is empty (Admin generated), we should allow multiple links, but we still need a DB entry.
	// We'll enforce the limit for everyone for simplicity and spam protection. Admin can manually delete via UI if needed.

	if reset, err := cd.queries.GetPasswordResetByUser(cd.ctx, db.GetPasswordResetByUserParams{
		UserID:    userID,
		CreatedAt: time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour),
	}); err == nil {
		// If existing reset is found
		if time.Since(reset.CreatedAt) < 24*time.Hour {
			// If recently created, reject (anti-spam)
			return "", ErrPasswordResetRecentlyRequested
		}
		// Else delete old one
		_ = cd.queries.SystemDeletePasswordReset(cd.ctx, reset.ID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get reset: %v", err)
		return "", fmt.Errorf("CoreData.CreatePasswordReset: get reset %w", err)
	}

	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("CoreData.CreatePasswordReset: rand %w", err)
	}
	code := hex.EncodeToString(buf[:])

	if err := cd.queries.CreatePasswordResetForUser(cd.ctx, db.CreatePasswordResetForUserParams{
		UserID: userID,
		Passwd: sql.NullString{String: hash, Valid: hash != ""},
		PasswdAlgorithm: sql.NullString{String: alg, Valid: alg != ""},
		VerificationCode: code,
	}); err != nil {
		return "", fmt.Errorf("CoreData.CreatePasswordReset: create reset %w", err)
	}

	// Insert into request queue (Audit/Log) - Only if it's a user request (hash provided)
	if hash != "" {
		reset, err := cd.queries.GetPasswordResetByUser(cd.ctx, db.GetPasswordResetByUserParams{
			UserID:    userID,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		})
		if err == nil {
			_, _ = cd.queries.AdminInsertRequestQueue(cd.ctx, db.AdminInsertRequestQueueParams{
				UsersIdusers: userID,
				ChangeTable:  "pending_passwords",
				ChangeField:  "password",
				ChangeRowID:  reset.ID,
				ChangeValue:  sql.NullString{String: "hidden", Valid: true},
			})
		}
	}

	return code, nil
}

// SignPasswordResetLink generates a signed URL for a password reset.
// It uses a dedicated key or a derived one. Here we use LinkSignKey for simplicity unless a new one is added.
func (cd *CoreData) SignPasswordResetLink(code string, expiry time.Duration) string {
	// We want to sign: code
	// And include expiry.
	// sign.Sign supports WithExpiry.

	exp := time.Now().Add(expiry)
	sig := sign.Sign("reset:"+code, cd.LinkSignKey, sign.WithExpiry(exp))

	// Add signature and expiry to query
	fullURL := fmt.Sprintf("/login/reset?code=%s", code)
	signedURL, _ := sign.AddQuerySig(fullURL, sig, sign.WithExpiry(exp))
	return signedURL
}

// VerifyPasswordResetLink verifies the signature of a password reset link.
// It does NOT verify the DB entry, just the link integrity and expiry.
// returns expiration time if valid
func (cd *CoreData) VerifyPasswordResetLink(code, sig, expiresStr string) (time.Time, error) {
	var opts []sign.SignOption
	var ts int64
	var expires time.Time

	if expiresStr != "" {
		var err error
		ts, err = strconv.ParseInt(expiresStr, 10, 64)
		if err == nil {
			expires = time.Unix(ts, 0)
			opts = append(opts, sign.WithExpiry(expires))
		}
	}

	data := "reset:" + code
	if err := sign.Verify(data, sig, cd.LinkSignKey, opts...); err != nil {
		return time.Time{}, err
	}

	if !expires.IsZero() && time.Now().After(expires) {
		return expires, errors.New("expired link")
	}

	return expires, nil
}


// VerifyPasswordReset validates a reset code and activates the new password hash when successful.
// code: verification code
// newHash: the new password (cleartext) provided by the user at verification time.
// optionalExpiry: if provided (from a signed link), it overrides the default DB-based expiry check.
func (cd *CoreData) VerifyPasswordReset(code string, newHash string, optionalExpiry *time.Time) error {
	if cd.queries == nil {
		return errors.New("missing queries")
	}

	var expiryLookback time.Time

	if optionalExpiry != nil {
		if !optionalExpiry.IsZero() && time.Now().After(*optionalExpiry) {
			return errors.New("link expired")
		}
		// If signed expiry is valid, we allow the DB entry to be older than the default config.
		// We pass the zero time to the query to disable the created_at check.
		expiryLookback = time.Time{}
	} else {
		// Use a reasonable lookback for created_at if expiry is not set
		expiryLookback = time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
	}

	reset, err := cd.queries.GetPasswordResetByCode(cd.ctx, db.GetPasswordResetByCodeParams{VerificationCode: code, CreatedAt: expiryLookback})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid code or expired")
		}
		return fmt.Errorf("CoreData.VerifyPasswordReset: get reset %w", err)
	}
	if _, err := cd.queries.GetLoginRoleForUser(cd.ctx, reset.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("approval is pending")
		}
		return fmt.Errorf("CoreData.VerifyPasswordReset: user role %w", err)
	}

	var passwordToStore string
	var algToStore string

	if reset.Passwd.Valid {
		// User requested reset with specific password
		if !VerifyPassword(newHash, reset.Passwd.String, reset.PasswdAlgorithm.String) {
			return errors.New("invalid password")
		}
		passwordToStore = reset.Passwd.String
		algToStore = reset.PasswdAlgorithm.String
	} else {
		// Admin generated link (no password preset)
		ph, alg, err := HashPassword(newHash)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		passwordToStore = ph
		algToStore = alg
	}

	if err := cd.queries.SystemMarkPasswordResetVerified(cd.ctx, reset.ID); err != nil {
		log.Printf("mark reset verified: %v", err)
	}
	if err := cd.queries.InsertPassword(cd.ctx, db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: passwordToStore, PasswdAlgorithm: sql.NullString{String: algToStore, Valid: true}}); err != nil {
		log.Printf("insert password: %v", err)
	}
	return nil
}
