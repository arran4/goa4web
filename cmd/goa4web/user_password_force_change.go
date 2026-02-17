package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
)

type userPasswordForceChangeCmd struct {
	*userCmd
	fs       *flag.FlagSet
	username string
	userID   int
}

func parseUserPasswordForceChangeCmd(parent *userCmd, args []string) (*userPasswordForceChangeCmd, error) {
	c := &userPasswordForceChangeCmd{userCmd: parent}
	fs := flag.NewFlagSet("force-change", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.username, "username", "", "Username")
	fs.IntVar(&c.userID, "user-id", 0, "User ID")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userPasswordForceChangeCmd) Run() error {
	ctx := c.rootCmd.Context()

	d, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	defer d.Close()

	queries := db.New(d)
	return forceChangePassword(ctx, queries, c.userID, c.username)
}

func forceChangePassword(ctx context.Context, queries *db.Queries, userID int, username string) error {
	uid, err := resolveUserID(ctx, queries, userID, username)
	if err != nil {
		return err
	}

	// Get username if we only have ID
	if username == "" {
		u, err := queries.SystemGetUserByID(ctx, uid)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		username = u.Username.String
	}

	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Errorf("rand: %w", err)
	}
	newPass := hex.EncodeToString(buf[:])

	hash, alg, err := auth.HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	if err := queries.InsertPassword(ctx, db.InsertPasswordParams{
		UsersIdusers:    uid,
		Passwd:          hash,
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
	}); err != nil {
		return fmt.Errorf("insert password: %w", err)
	}

	if _, err := queries.SystemDeletePasswordResetsByUser(ctx, uid); err != nil {
		return fmt.Errorf("clear resets: %w", err)
	}

	fmt.Printf("Password for user %q (ID: %d) changed to: %s\n", username, uid, newPass)
	return nil
}

// Usage prints command usage information with examples.
func (c *userPasswordForceChangeCmd) Usage() {
	executeUsage(c.fs.Output(), "user_password_force_change_usage.txt", c)
}

func (c *userPasswordForceChangeCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordForceChangeCmd)(nil)
