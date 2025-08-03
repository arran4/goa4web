package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
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
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(db)

	var rows []*db.SystemListUserInfoRow
	if c.showAdmin || c.showCreated || c.jsonOut {
		rows, err = queries.SystemListUserInfo(ctx)
	} else {
		// fall back to basic user list when no extra columns requested
		basic, err2 := queries.AdminListAllUsers(ctx)
		if err2 != nil {
			return fmt.Errorf("list users: %w", err2)
		}
		for _, u := range basic {
			fmt.Printf("%d\t%s\t%s\n", u.Idusers, u.Username.String, u.Email)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	if c.jsonOut {
		out := make([]map[string]interface{}, 0, len(rows))
		for _, u := range rows {
			item := map[string]interface{}{
				"id":       u.Idusers,
				"username": u.Username.String,
				"email":    u.Email,
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
		fmt.Printf("%d\t%s\t%s", u.Idusers, u.Username.String, u.Email)
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
