package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// userListCmd implements "user list".
type userListCmd struct {
	*userCmd
	fs          *flag.FlagSet
	showAdmin   bool
	showCreated bool
	jsonOut     bool
}

func parseUserListCmd(parent *userCmd, args []string) (*userListCmd, error) {
	c := &userListCmd{userCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&c.showAdmin, "admin", false, "include admin status")
		fs.BoolVar(&c.showCreated, "created", false, "include creation date")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	rows, err := queries.SystemListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	emailRows, err := queries.GetVerifiedUserEmails(ctx)
	if err != nil {
		return fmt.Errorf("list user emails: %w", err)
	}
	emailsByUser := db.EmailsByUserID(emailRows)

	if c.jsonOut {
		out := make([]map[string]interface{}, 0, len(rows))
		for _, u := range rows {
			userEmails := emailsByUser[u.Idusers]
			item := map[string]interface{}{
				"id":       u.Idusers,
				"username": u.Username.String,
				"email":    db.PrimaryEmail(userEmails),
				"emails":   userEmails,
			}
			if c.showAdmin || c.jsonOut {
				item["admin"] = u.Admin
			}
			if c.showCreated || c.jsonOut {
				if t, ok := u.CreatedAt.(sql.NullTime); ok && t.Valid {
					item["created_at"] = t.Time.Format(time.RFC3339)
				}
			}
			out = append(out, item)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	for _, u := range rows {
		userEmails := emailsByUser[u.Idusers]
		fmt.Printf("%d\t%s\t%s", u.Idusers, u.Username.String, strings.Join(userEmails, ","))
		if c.showAdmin {
			fmt.Printf("\t%t", u.Admin)
		}
		if c.showCreated {
			if t, ok := u.CreatedAt.(sql.NullTime); ok && t.Valid {
				fmt.Printf("\t%s", t.Time.Format(time.RFC3339))
			} else {
				fmt.Printf("\t-")
			}
		}
		fmt.Println()
	}
	return nil
}
