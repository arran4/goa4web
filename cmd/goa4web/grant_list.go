package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// grantListCmd implements "grant list".
type grantListCmd struct {
	*grantCmd
	fs       *flag.FlagSet
	filter   string
	userID   int
	uid      int
	username string
	roleID   int
	rid      int
	roleName string
}

func parseGrantListCmd(parent *grantCmd, args []string) (*grantListCmd, error) {
	c := &grantListCmd{grantCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.StringVar(&c.filter, "filter", "roles", "Filter by 'roles', 'users', or 'both'")
	c.fs.IntVar(&c.userID, "user-id", 0, "Filter by user ID")
	c.fs.IntVar(&c.uid, "uid", 0, "Filter by user ID (alias for -user-id)")
	c.fs.StringVar(&c.username, "username", "", "Filter by username")
	c.fs.IntVar(&c.roleID, "role-id", 0, "Filter by role ID")
	c.fs.IntVar(&c.rid, "rid", 0, "Filter by role ID (alias for -role-id)")
	c.fs.StringVar(&c.roleName, "role-name", "", "Filter by role name")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *grantListCmd) Run() error {
	q, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("querier: %w", err)
	}

	if c.uid != 0 {
		c.userID = c.uid
	}
	if c.rid != 0 {
		c.roleID = c.rid
	}

	hasUserFilter := c.userID != 0 || c.username != ""
	hasRoleFilter := c.roleID != 0 || c.roleName != ""

	if hasUserFilter && hasRoleFilter {
		return fmt.Errorf("cannot combine user-specific filters (--user-id, --username) with role-specific filters (--role-id, --role-name)")
	}

	params := db.ListGrantsExtendedParams{}
	if c.username != "" {
		params.Username = sql.NullString{String: c.username, Valid: true}
	}
	if c.roleName != "" {
		params.RoleName = sql.NullString{String: c.roleName, Valid: true}
	}
	if c.userID != 0 {
		params.UserID = sql.NullInt32{Int32: int32(c.userID), Valid: true}
	}
	if c.roleID != 0 {
		params.RoleID = sql.NullInt32{Int32: int32(c.roleID), Valid: true}
	}

	filterType := c.filter
	if hasUserFilter {
		filterType = "users"
	} else if hasRoleFilter {
		filterType = "roles"
	}

	switch filterType {
	case "roles":
		params.OnlyRoles = true
	case "users":
		params.OnlyUsers = true
	case "both":
	default:
		return fmt.Errorf("invalid filter: %q", c.filter)
	}

	ctx := context.Background()
	rows, err := q.ListGrantsExtended(ctx, params)
	if err != nil {
		return fmt.Errorf("list grants: %w", err)
	}
	if err := printGrantsTable(c.fs.Output(), rows); err != nil {
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
