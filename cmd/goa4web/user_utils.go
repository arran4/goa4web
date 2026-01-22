package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
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

func getResetURL(
	ctx context.Context,
	queries *db.Queries,
	cfg *config.RuntimeConfig,
	userID int,
	username string,
) (string, string, int32, error) {
	if username == "" && userID == 0 {
		return "", "", 0, fmt.Errorf("either --username or --user-id must be provided")
	}
	if username != "" && userID != 0 {
		return "", "", 0, fmt.Errorf("only one of --username or --user-id can be provided")
	}

	var user struct {
		ID       int32
		Username string
	}

	if userID != 0 {
		u, err := queries.SystemGetUserByID(ctx, int32(userID))
		if err != nil {
			return "", "", 0, fmt.Errorf("failed to get user by id %d: %w", userID, err)
		}
		user.ID = u.Idusers
		user.Username = u.Username.String
	} else {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: username, Valid: true})
		if err != nil {
			return "", "", 0, fmt.Errorf("failed to get user by username %q: %w", username, err)
		}
		user.ID = u.Idusers
		user.Username = u.Username.String
	}

	key, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cfg.LinkSignSecret, cfg.LinkSignSecretFile)
	if err != nil {
		return "", "", 0, fmt.Errorf("link sign secret: %w", err)
	}

	code := generateVerificationCode()
	if err := queries.CreatePasswordResetForUser(ctx, db.CreatePasswordResetForUserParams{
		UserID:           user.ID,
		Passwd:           "magic-link",
		PasswdAlgorithm:  "magic",
		VerificationCode: code,
	}); err != nil {
		return "", "", 0, fmt.Errorf("create db token: %w", err)
	}

	targetURL := fmt.Sprintf("/user/%d/reset?code=%s", user.ID, code)
	data := "link:" + targetURL

	duration := time.Hour * 24
	opts := []sign.SignOption{
		sign.WithExpiry(time.Now().Add(duration)),
	}

	sig := sign.Sign(data, key, opts...)
	base, err := url.Parse(cfg.HTTPHostname)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to parse hostname: %w", err)
	}
	signedURL, err := sign.AddQuerySig(base.JoinPath(targetURL).String(), sig, opts...)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to add query signature: %w", err)
	}
	return signedURL, user.Username, user.ID, nil
}

func generatePasswordReset(
	ctx context.Context,
	queries *db.Queries,
	cfg *config.RuntimeConfig,
	userID int,
	username string,
) error {
	signedURL, uName, uid, err := getResetURL(ctx, queries, cfg, userID, username)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully generated password reset link for user %q (ID: %d).\n", uName, uid)
	fmt.Println("\nPlease provide the following URL to the user.")
	fmt.Println("This link is valid for 24 hours.")
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Printf("Reset URL: %s\n", signedURL)
	fmt.Println("--------------------------------------------------------------------------")

	return nil
}
