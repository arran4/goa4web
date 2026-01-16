package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
)

// userAddCmd implements the "user add" command.
type userAddCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Email    string
	Password string
	Admin    bool
}

func parseUserAddCmd(parent *userCmd, args []string) (*userAddCmd, error) {
	c := &userAddCmd{userCmd: parent}
	fs, _, err := parseFlags("add", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "The username for the new user. This is a required field.")
		fs.StringVar(&c.Email, "email", "", "The email address for the new user. This is an optional field.")
		fs.StringVar(&c.Password, "password", "", "The password for the new user. If not provided, you will be prompted to enter it.")
		fs.BoolVar(&c.Admin, "admin", false, "If set, the new user will be granted administrator rights.")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userAddCmd) Usage() {
	executeUsage(c.fs.Output(), "user_add_usage.txt", c)
}

func (c *userAddCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userAddCmd)(nil)

func (c *userAddCmd) Run() error {
	pw := c.Password
	if pw == "" {
		var err error
		if pw, err = promptPassword(); err != nil {
			return fmt.Errorf("prompt password: %w", err)
		}
	}
	return createUser(c.userCmd.rootCmd, c.Username, c.Email, pw, c.Admin)
}

func createUser(root *rootCmd, username, email, password string, admin bool) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password required")
	}
	conn, err := root.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	hash, alg, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	id, err := queries.SystemInsertUser(ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("insert user: %w", err)
	}
	if email != "" {
		if err := queries.InsertUserEmail(ctx, db.InsertUserEmailParams{UserID: int32(id), Email: email, VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, LastVerificationCode: sql.NullString{}, NotificationPriority: 100}); err != nil {
			log.Printf("insert user email: %v", err)
		}
	}
	if err := queries.InsertPassword(ctx, db.InsertPasswordParams{UsersIdusers: int32(id), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: alg != ""}}); err != nil {
		return fmt.Errorf("insert password: %w", err)
	}
	if !admin {
		if err := queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
			UsersIdusers: int32(id),
			Name:         "user",
		}); err != nil {
			return fmt.Errorf("grant role: %w", err)
		}
	} else {
		if _, err := queries.GetAdministratorUserRole(ctx, int32(id)); err == nil {
			if root.Verbosity > 0 {
				fmt.Printf("%s already administrator\n", username)
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check admin: %w", err)
		} else if err := queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
			UsersIdusers: int32(id),
			Name:         "administrator",
		}); err != nil {
			return fmt.Errorf("grant admin: %w", err)
		}
	}
	if root.Verbosity > 0 {
		fmt.Printf("created user %s (id %d)\n", username, id)
	}
	return nil
}
