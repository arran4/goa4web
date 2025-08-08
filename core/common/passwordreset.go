package common

import (
	"crypto/md5"
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

// VerifyPasswordReset validates a reset code and activates the new password
// hash when successful.
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
