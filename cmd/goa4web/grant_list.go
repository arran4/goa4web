package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// grantListCmd implements "grant list".
type grantListCmd struct {
	*grantCmd
	fs *flag.FlagSet

	filter   string
	uid      int64
	username string
	rid      int64
	rolename string
}

func parseGrantListCmd(parent *grantCmd, args []string) (*grantListCmd, error) {
	c := &grantListCmd{grantCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.StringVar(&c.filter, "filter", "roles", "Filter grants by 'roles', 'users', or 'both'")
	c.fs.Int64Var(&c.uid, "uid", 0, "Filter by user ID")
	c.fs.Int64Var(&c.uid, "user-id", 0, "Filter by user ID")
	c.fs.StringVar(&c.username, "username", "", "Filter by username")
	c.fs.Int64Var(&c.rid, "rid", 0, "Filter by role ID")
	c.fs.Int64Var(&c.rid, "role-id", 0, "Filter by role ID")
	c.fs.StringVar(&c.rolename, "role-name", "", "Filter by role name")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *grantListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	q := db.New(conn)
	filter := c.filter
	if c.uid != 0 || c.username != "" {
		if c.rid != 0 || c.rolename != "" {
			filter = "both"
		} else {
			filter = "users"
		}
	} else if c.rid != 0 || c.rolename != "" {
		filter = "roles"
	}
	params := db.ListGrantsExtendedParams{
		Filter:   filter,
		UserID:   c.uid,
		Username: c.username,
		RoleID:   c.rid,
		RoleName: c.rolename,
	}
	rows, err := q.ListGrantsExtended(ctx, params)
	if err != nil {
		return fmt.Errorf("list grants: %w", err)
	}
	if err := printGrantsTable(os.Stdout, rows); err != nil {
		return fmt.Errorf("printing grants table: %w", err)
	}
	return nil
}

func printGrantsTable(out io.Writer, rows []*db.ListGrantsExtendedRow) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive")
	for _, g := range rows {
		target := "Everyone"
		if g.RoleName.Valid {
			target = fmt.Sprintf("Role: %s", g.RoleName.String)
		} else if g.Username.Valid {
			target = fmt.Sprintf("User: %s", g.Username.String)
		} else if g.RoleID.Valid {
			target = fmt.Sprintf("Role ID: %d", g.RoleID.Int32)
		} else if g.UserID.Valid {
			target = fmt.Sprintf("User ID: %d", g.UserID.Int32)
		}

		item := "-"
		if g.Item.Valid && g.Item.String != "" {
			item = g.Item.String
		}

		scope := "All"
		if g.ItemID.Valid {
			scope = fmt.Sprintf("ID: %d", g.ItemID.Int32)
		} else if g.ItemRule.Valid && g.ItemRule.String != "" {
			scope = fmt.Sprintf("Rule: %s", g.ItemRule.String)
		}

		ruleType := g.RuleType
		if ruleType == "" {
			ruleType = "-"
		}

		active := "Yes"
		if !g.Active {
			active = "No"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", g.ID, g.Section, item, g.Action, ruleType, target, scope, active)
	}
	return w.Flush()
}
