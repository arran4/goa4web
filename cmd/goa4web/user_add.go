package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userAddCmd implements the "user add" command.
type userAddCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Email    string
	Password string
	Admin    bool
	args     []string
}

func parseUserAddCmd(parent *userCmd, args []string) (*userAddCmd, error) {
	c := &userAddCmd{userCmd: parent}
	fs, rest, err := parseFlags("add", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Email, "email", "", "email address")
		fs.StringVar(&c.Password, "password", "", "password (leave empty to prompt)")
		fs.BoolVar(&c.Admin, "admin", false, "grant administrator rights")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

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
	db, err := root.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	hash, alg, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	res, err := queries.DB().ExecContext(ctx,
		"INSERT INTO users (username) VALUES (?)",
		username,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	if email != "" {
		_ = queries.InsertUserEmail(ctx, dbpkg.InsertUserEmailParams{UserID: int32(id), Email: email, VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, LastVerificationCode: sql.NullString{}, NotificationPriority: 100})
	}
	if err := queries.InsertPassword(ctx, dbpkg.InsertPasswordParams{UsersIdusers: int32(id), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: alg != ""}}); err != nil {
		return fmt.Errorf("insert password: %w", err)
	}
	if admin {
		if _, err := queries.GetAdministratorUserRole(ctx, int32(id)); err == nil {
			if root.Verbosity > 0 {
				fmt.Printf("%s already administrator\n", username)
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check admin: %w", err)
		} else if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{
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

// hashPassword creates a PBKDF2-SHA256 hash and algorithm string.
func hashPassword(pw string) (string, string, error) {
	const iterations = 10000
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", "", err
	}
	hash := pbkdf2.Key([]byte(pw), salt[:], iterations, 32, sha256.New)
	alg := fmt.Sprintf("pbkdf2-sha256:%d:%x", iterations, salt)
	return hex.EncodeToString(hash), alg, nil
}
