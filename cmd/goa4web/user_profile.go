package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userProfileCmd implements "user profile" to show user details.
type userProfileCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
}

func parseUserProfileCmd(parent *userCmd, args []string) (*userProfileCmd, error) {
	c := &userProfileCmd{userCmd: parent}
	fs, _, err := parseFlags("profile", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userProfileCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if c.ID == 0 {
		u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	u, err := queries.GetUserById(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	fmt.Printf("ID: %d\nUsername: %s\n", c.ID, u.Username.String)
	emails, _ := queries.GetUserEmailsByUserID(ctx, int32(c.ID))
	for _, e := range emails {
		fmt.Printf("Email: %s verified:%t priority:%d\n", e.Email, e.VerifiedAt.Valid, e.NotificationPriority)
	}
	comments, _ := queries.ListAdminUserComments(ctx, int32(c.ID))
	if len(comments) > 0 {
		fmt.Println("Admin comments:")
		for _, cm := range comments {
			fmt.Printf("%s %s\n", cm.CreatedAt.Format("2006-01-02 15:04"), cm.Comment)
		}
	}
	return nil
}
