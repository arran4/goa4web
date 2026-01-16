package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userAddRoleCmd implements the "user add-role" command.
type userAddRoleCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Role     string
	RoleID   int
}

func parseUserAddRoleCmd(parent *userCmd, args []string) (*userAddRoleCmd, error) {
	c := &userAddRoleCmd{userCmd: parent}
	fs, _, err := parseFlags("add-role", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Role, "role", "", "role name")
		fs.IntVar(&c.RoleID, "role-id", 0, "role ID")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userAddRoleCmd) Usage() {
	executeUsage(c.fs.Output(), "user_add_role_usage.txt", c)
}

func (c *userAddRoleCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userAddRoleCmd)(nil)

func (c *userAddRoleCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	if c.Role == "" && c.RoleID == 0 {
		return fmt.Errorf("either --role or --role-id required")
	}
	if c.Role != "" && c.RoleID != 0 {
		return fmt.Errorf("cannot specify both --role and --role-id")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	// Get user
	u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	var roleID int32
	var roleName string

	// Determine role ID and name
	if c.RoleID != 0 {
		roleID = int32(c.RoleID)
		// Verify role exists and get its name
		role, err := queries.AdminGetRoleByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("role ID %d does not exist", c.RoleID)
			}
			return fmt.Errorf("get role: %w", err)
		}
		roleName = role.Name
	} else {
		roleName = c.Role
		// Verify role exists and get its ID
		role, err := queries.GetRoleByName(ctx, c.Role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("role %s does not exist", c.Role)
			}
			return fmt.Errorf("get role: %w", err)
		}
		roleID = role.ID
	}

	c.rootCmd.Verbosef("adding role %s (ID: %d) to %s", roleName, roleID, c.Username)

	// Check if user already has this role
	if _, err := queries.AdminGetRoleByNameForUser(ctx, db.AdminGetRoleByNameForUserParams{
		UsersIdusers: u.Idusers,
		Name:         roleName,
	}); err == nil {
		c.rootCmd.Verbosef("%s already has role %s", c.Username, roleName)
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check role: %w", err)
	}

	// Add the role
	if err := queries.SystemCreateUserRoleByID(ctx, db.SystemCreateUserRoleByIDParams{
		UsersIdusers: u.Idusers,
		RoleID:       roleID,
	}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}

	c.rootCmd.Infof("added role %s (ID: %d) to %s", roleName, roleID, c.Username)
	return nil
}
