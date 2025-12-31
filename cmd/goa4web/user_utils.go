package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
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
