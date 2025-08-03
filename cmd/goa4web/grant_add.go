package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// grantAddCmd implements "grant add".
type grantAddCmd struct {
	*grantCmd
	fs      *flag.FlagSet
	User    int
	Role    string
	Section string
	Item    string
	Action  string
	ItemID  int
}

func parseGrantAddCmd(parent *grantCmd, args []string) (*grantAddCmd, error) {
	c := &grantAddCmd{grantCmd: parent}
	c.fs = newFlagSet("add")
	c.fs.IntVar(&c.User, "user-id", 0, "user id")
	c.fs.StringVar(&c.Role, "role", "", "role name")
	c.fs.StringVar(&c.Section, "section", "", "section name")
	c.fs.StringVar(&c.Item, "item", "", "item name")
	c.fs.StringVar(&c.Action, "action", "", "action name")
	c.fs.IntVar(&c.ItemID, "item-id", 0, "item id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *grantAddCmd) Run() error {
	if c.Section == "" || c.Action == "" {
		return fmt.Errorf("section and action required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	q := db.New(conn)
	_, err = q.AdminCreateGrant(ctx, db.AdminCreateGrantParams{
		UserID:   sql.NullInt32{Int32: int32(c.User), Valid: c.User != 0},
		RoleID:   sql.NullInt32{},
		Section:  c.Section,
		Item:     sql.NullString{String: c.Item, Valid: c.Item != ""},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: int32(c.ItemID), Valid: c.ItemID != 0},
		ItemRule: sql.NullString{},
		Action:   c.Action,
		Extra:    sql.NullString{},
	})
	if err != nil {
		return fmt.Errorf("create grant: %w", err)
	}
	return nil
}
