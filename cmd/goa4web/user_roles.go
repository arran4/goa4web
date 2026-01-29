package main

import (
	"context"
	"flag"
	"fmt"
	"sort"

	"github.com/arran4/goa4web/internal/db"
)

// userRolesCmd implements "user roles".
type userRolesCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserRolesCmd(parent *userCmd, args []string) (*userRolesCmd, error) {
	c := &userRolesCmd{userCmd: parent}
	c.fs = newFlagSet("roles")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userRolesCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		return c.runList()
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "sync":
		cmd, err := parseUserRolesSyncCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown roles command %q", args[0])
	}
}

func (c *userRolesCmd) runList() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.ListUsersWithRoles(ctx)
	if err != nil {
		return fmt.Errorf("list users with roles: %w", err)
	}
	for _, r := range rows {
		roleList := ""
		if r.Roles.Valid {
			roleList = r.Roles.String
		}
		fmt.Printf("%s\t%s\n", r.Username.String, roleList)
	}
	return nil
}

func (c *userRolesCmd) Usage() {
	executeUsage(c.fs.Output(), "user_roles_usage.txt", c)
}

func (c *userRolesCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userRolesCmd)(nil)

// userRolesSyncCmd implements "user roles sync".
type userRolesSyncCmd struct {
	*userRolesCmd
	fs       *flag.FlagSet
	template string
	dryRun   bool
}

func parseUserRolesSyncCmd(parent *userRolesCmd, args []string) (*userRolesSyncCmd, error) {
	c := &userRolesSyncCmd{userRolesCmd: parent}
	fs, _, err := parseFlags("sync", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.template, "template", "", "role template name")
		fs.BoolVar(&c.dryRun, "dry-run", false, "print planned changes without applying them")
	})
	if err != nil {
		return nil, err
	}
	if c.template == "" {
		return nil, fmt.Errorf("template is required")
	}
	c.fs = fs
	return c, nil
}

func (c *userRolesSyncCmd) Run() error {
	template, ok := roleTemplates[c.template]
	if !ok {
		return fmt.Errorf("template %q not found", c.template)
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	roles, err := queries.AdminListRoles(ctx)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}

	roleIndex := make(map[string]*db.Role, len(roles))
	for _, role := range roles {
		roleIndex[role.Name] = role
	}

	templateRoleIDs := make(map[string]int32, len(template.Roles))
	templateRoles := make(map[string]RoleDef, len(template.Roles))
	var templateAdminRoles []string
	var templateLoginRoles []string
	var templateNonLoginRoles []string
	for _, role := range template.Roles {
		templateRoles[role.Name] = role
		dbRole, ok := roleIndex[role.Name]
		if !ok {
			return fmt.Errorf("template role %q not found in database", role.Name)
		}
		templateRoleIDs[role.Name] = dbRole.ID
		if role.IsAdmin {
			templateAdminRoles = append(templateAdminRoles, role.Name)
		}
		if role.CanLogin && !role.IsAdmin {
			templateLoginRoles = append(templateLoginRoles, role.Name)
		}
		if !role.CanLogin && !role.IsAdmin {
			templateNonLoginRoles = append(templateNonLoginRoles, role.Name)
		}
	}

	users, err := queries.ListUsersWithRoles(ctx)
	if err != nil {
		return fmt.Errorf("list users with roles: %w", err)
	}

	for _, user := range users {
		username := user.Username.String
		if !user.Username.Valid {
			username = fmt.Sprintf("user-%d", user.Idusers)
		}
		perms, err := queries.GetPermissionsByUserID(ctx, user.Idusers)
		if err != nil {
			return fmt.Errorf("list roles for user %d: %w", user.Idusers, err)
		}

		currentRoles := make(map[string]int32, len(perms))
		desiredRoles := make(map[string]struct{})
		for _, perm := range perms {
			currentRoles[perm.Name] = perm.IduserRoles
			if _, ok := templateRoles[perm.Name]; ok {
				desiredRoles[perm.Name] = struct{}{}
				continue
			}
			roleInfo, ok := roleIndex[perm.Name]
			if !ok {
				return fmt.Errorf("role %q not found in database", perm.Name)
			}
			mapped := mapTemplateRole(roleInfo, templateAdminRoles, templateLoginRoles, templateNonLoginRoles)
			if mapped != "" {
				desiredRoles[mapped] = struct{}{}
			}
		}

		addRoles, removeRoles := diffRoleSets(currentRoles, desiredRoles)
		if len(addRoles) == 0 && len(removeRoles) == 0 {
			continue
		}

		if c.dryRun {
			fmt.Printf("%s (id %d):\n", username, user.Idusers)
			for _, role := range addRoles {
				fmt.Printf("  + %s\n", role)
			}
			for _, role := range removeRoles {
				fmt.Printf("  - %s\n", role)
			}
			continue
		}

		for _, role := range addRoles {
			roleID, ok := templateRoleIDs[role]
			if !ok {
				return fmt.Errorf("role %q missing from template role IDs", role)
			}
			if err := queries.SystemCreateUserRoleByID(ctx, db.SystemCreateUserRoleByIDParams{
				UsersIdusers: user.Idusers,
				RoleID:       roleID,
			}); err != nil {
				return fmt.Errorf("add role %s for %s: %w", role, username, err)
			}
			c.rootCmd.Infof("added role %s (ID: %d) to %s", role, roleID, username)
		}

		for _, role := range removeRoles {
			roleID, ok := currentRoles[role]
			if !ok {
				return fmt.Errorf("role %q missing from user roles", role)
			}
			if err := queries.AdminDeleteUserRole(ctx, roleID); err != nil {
				return fmt.Errorf("remove role %s for %s: %w", role, username, err)
			}
			c.rootCmd.Infof("removed role %s from %s", role, username)
		}
	}

	return nil
}

func mapTemplateRole(role *db.Role, adminRoles, loginRoles, nonLoginRoles []string) string {
	if role.IsAdmin && len(adminRoles) == 1 {
		return adminRoles[0]
	}
	if role.CanLogin && !role.IsAdmin && len(loginRoles) == 1 {
		return loginRoles[0]
	}
	if !role.CanLogin && !role.IsAdmin && len(nonLoginRoles) == 1 {
		return nonLoginRoles[0]
	}
	return ""
}

func diffRoleSets(current map[string]int32, desired map[string]struct{}) ([]string, []string) {
	var addRoles []string
	var removeRoles []string
	for role := range desired {
		if _, ok := current[role]; !ok {
			addRoles = append(addRoles, role)
		}
	}
	for role := range current {
		if _, ok := desired[role]; !ok {
			removeRoles = append(removeRoles, role)
		}
	}
	sort.Strings(addRoles)
	sort.Strings(removeRoles)
	return addRoles, removeRoles
}

func (c *userRolesSyncCmd) Usage() {
	executeUsage(c.fs.Output(), "user_roles_usage.txt", c)
}

func (c *userRolesSyncCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userRolesSyncCmd)(nil)
