package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userListCmd implements "user list".
type userListCmd struct {
	*userCmd
	fs          *flag.FlagSet
	showAdmin   bool
	showCreated bool
	jsonOut     bool
	args        []string
}

func parseUserListCmd(parent *userCmd, args []string) (*userListCmd, error) {
	c := &userListCmd{userCmd: parent}
	fs, rest, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&c.showAdmin, "admin", false, "include admin status")
		fs.BoolVar(&c.showCreated, "created", false, "include creation date")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

func (c *userListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)

	var rows []*dbpkg.UserInfoRow
	if c.showAdmin || c.showCreated || c.jsonOut {
		rows, err = queries.ListUserInfo(ctx)
	} else {
		// fall back to basic user list when no extra columns requested
		basic, err2 := queries.AllUsers(ctx)
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
				"id":       u.ID,
				"username": u.Username.String,
				"email":    u.Email.String,
			}
			if c.showAdmin || c.jsonOut {
				item["admin"] = u.Admin
			}
			if c.showCreated || c.jsonOut {
				if u.CreatedAt.Valid {
					item["created_at"] = u.CreatedAt.Time.Format(time.RFC3339)
				}
			}
			out = append(out, item)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	for _, u := range rows {
		fmt.Printf("%d\t%s\t%s", u.ID, u.Username.String, u.Email.String)
		if c.showAdmin {
			fmt.Printf("\t%t", u.Admin)
		}
		if c.showCreated {
			if u.CreatedAt.Valid {
				fmt.Printf("\t%s", u.CreatedAt.Time.Format(time.RFC3339))
			} else {
				fmt.Printf("\t-")
			}
		}
		fmt.Println()
	}
	return nil
}
