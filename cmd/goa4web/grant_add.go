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
	fs       *flag.FlagSet
	UserID   int
	UserName string
	Role     string
	RoleID   int
	Section  string
	Item     string
	Action   string
	ItemID   int
}

func parseGrantAddCmd(parent *grantCmd, args []string) (*grantAddCmd, error) {
	c := &grantAddCmd{grantCmd: parent}
	c.fs = newFlagSet("add")
	c.fs.IntVar(&c.UserID, "user-id", 0, "user id")
	c.fs.StringVar(&c.UserName, "user-name", "", "user name")
	c.fs.StringVar(&c.Role, "role", "", "role name")
	c.fs.IntVar(&c.RoleID, "role-id", 0, "role id")
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
	if c.UserID != 0 && c.UserName != "" {
		return fmt.Errorf("cannot specify both -user-id and -user-name")
	}
	if c.RoleID != 0 && c.Role != "" {
		return fmt.Errorf("cannot specify both -role-id and -role")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	q := db.New(conn)

	userID := c.UserID
	if c.UserName != "" {
		user, err := q.SystemGetUserByUsername(ctx, sql.NullString{String: c.UserName, Valid: c.UserName != ""})
		if err != nil {
			return fmt.Errorf("get user by username %q: %w", c.UserName, err)
		}
		userID = int(user.Idusers)
		c.rootCmd.Infof("Found user %s with ID %d", user.Username.String, userID)
	}

	roleID := c.RoleID
	if c.Role != "" {
		role, err := q.GetRoleByName(ctx, c.Role)
		if err != nil {
			return fmt.Errorf("get role by name %q: %w", c.Role, err)
		}
		roleID = int(role.ID)
		c.rootCmd.Infof("Found role %s with ID %d", role.Name, roleID)
	}

	if userID == 0 && roleID == 0 {
		return fmt.Errorf("either a user or a role must be specified")
	}
	if userID != 0 && roleID != 0 {
		return fmt.Errorf("cannot specify both a user and a role for the same grant")
	}

	var userLog, roleLog string
	if userID != 0 {
		if c.UserName == "" {
			user, err := q.SystemGetUserByID(ctx, int32(userID))
			if err != nil {
				return fmt.Errorf("get user by id %d: %w", userID, err)
			}
			c.UserName = user.Username.String
		}
		userLog = fmt.Sprintf("user %s (%d)", c.UserName, userID)
	}
	if roleID != 0 {
		if c.Role == "" {
			role, err := q.AdminGetRoleByID(ctx, int32(roleID))
			if err != nil {
				return fmt.Errorf("get role by id %d: %w", roleID, err)
			}
			c.Role = role.Name
		}
		roleLog = fmt.Sprintf("role %s (%d)", c.Role, roleID)
	}

	c.rootCmd.Infof("Adding grant for %s%s to %s/%s/%d action %s", userLog, roleLog, c.Section, c.Item, c.ItemID, c.Action)

	_, err = q.AdminCreateGrant(ctx, db.AdminCreateGrantParams{
		UserID:   sql.NullInt32{Int32: int32(userID), Valid: userID != 0},
		RoleID:   sql.NullInt32{Int32: int32(roleID), Valid: roleID != 0},
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

	c.rootCmd.Infof("Grant added successfully.")
	return nil
}
