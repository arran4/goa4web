package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	dbm "github.com/arran4/goa4web/internal/db"
	"golang.org/x/crypto/pbkdf2"
)

func resolveUserID(ctx context.Context, queries *db.Queries, userID int, username string) (int32, error) {
	if userID != 0 {
		return int32(userID), nil
	}
	if username != "" {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: username, Valid: true})
		if err != nil {
			return 0, fmt.Errorf("lookup username %q: %w", username, err)
		}
		return int32(u.Idusers), nil
	}
	return 0, fmt.Errorf("missing -user-id or -username")
}

func generateVerificationCode() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}

func generatePasswordReset(
	ctx context.Context,
	queries *db.Queries,
	cfg *config.RuntimeConfig,
	userID int,
	username string,
) error {
	if username == "" && userID == 0 {
		return fmt.Errorf("either --username or --user-id must be provided")
	}
	if username != "" && userID != 0 {
		return fmt.Errorf("only one of --username or --user-id can be provided")
	}

	var user struct {
		ID       int32
		Username string
	}

	if userID != 0 {
		u, err := queries.SystemGetUserByID(ctx, int32(userID))
		if err != nil {
			return fmt.Errorf("failed to get user by id %d: %w", userID, err)
		}
		user.ID = u.Idusers
		user.Username = u.Username.String
	} else {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: username, Valid: true})
		if err != nil {
			return fmt.Errorf("failed to get user by username %q: %w", username, err)
		}
		user.ID = u.Idusers
		user.Username = u.Username.String
	}

	var codeBuf [16]byte
	if _, err := rand.Read(codeBuf[:]); err != nil {
		return fmt.Errorf("failed to generate random code: %w", err)
	}
	code := hex.EncodeToString(codeBuf[:])

	var tempPassBuf [12]byte
	if _, err := rand.Read(tempPassBuf[:]); err != nil {
		return fmt.Errorf("failed to generate random password: %w", err)
	}
	tempPassword := hex.EncodeToString(tempPassBuf[:])

	var salt [8]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	iterations := 600000
	hash := pbkdf2.Key([]byte(tempPassword), salt[:], iterations, 32, sha256.New)
	alg := fmt.Sprintf("pbkdf2-sha256:%d:%s", iterations, hex.EncodeToString(salt[:]))

	params := dbm.CreatePasswordResetForUserParams{
		UserID:           user.ID,
		Passwd:           hex.EncodeToString(hash),
		PasswdAlgorithm:  alg,
		VerificationCode: code,
	}

	if err := queries.CreatePasswordResetForUser(ctx, params); err != nil {
		return fmt.Errorf("failed to create password reset in database: %w", err)
	}

	resetURL := fmt.Sprintf("%s/password-reset?code=%s", cfg.HTTPHostname, code)

	fmt.Printf("Successfully generated password reset link for user %q (ID: %d).\n", user.Username, user.ID)
	fmt.Println("\nPlease provide the following temporary password and URL to the user.")
	fmt.Println("The user will be required to enter the temporary password to set their new password.")
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Printf("Temporary Password: %s\n", tempPassword)
	fmt.Printf("Reset URL:          %s\n", resetURL)
	fmt.Println("--------------------------------------------------------------------------")

	return nil
}
