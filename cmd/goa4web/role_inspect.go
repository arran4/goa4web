package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// roleInspectCmd implements the "role inspect" subcommand.
type roleInspectCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	role string
}

func parseRoleInspectCmd(parent *roleCmd, args []string) (*roleInspectCmd, error) {
	c := &roleInspectCmd{roleCmd: parent}
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.role, "role", "", "The role to inspect.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return c, nil
}

func (c *roleInspectCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	r, err := q.GetRoleByName(ctx, c.role)
	if err != nil {
		return fmt.Errorf("failed to get role %q: %w", c.role, err)
	}

	fmt.Printf("Role: %s (ID: %d)\n", r.Name, r.ID)
	fmt.Printf("  Can Login: %v\n", r.CanLogin)
	fmt.Printf("  Is Admin: %v\n", r.IsAdmin)
	fmt.Printf("  Private Labels: %v\n", r.PrivateLabels)

	if r.PublicProfileAllowedAt.Valid {
		fmt.Printf("  Public Profile Allowed At: %s\n", r.PublicProfileAllowedAt.Time)
	}

	// List Users
	users, err := q.AdminListUsersByRoleID(ctx, r.ID)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	if len(users) > 0 {
		fmt.Println("Users:")
		for _, u := range users {
			fmt.Printf("  - %s (ID: %d)\n", u.Username.String, u.Idusers)
		}
	} else {
		fmt.Println("Users: None")
	}

	// List Grants
	grants, err := q.AdminListGrantsByRoleID(ctx, sql.NullInt32{Int32: r.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to list grants: %w", err)
	}

	if len(grants) > 0 {
		fmt.Println("Grants:")
		// Heuristic: Check against other roles to see if this role includes all their grants
		matches := checkRoleInclusion(q, ctx, grants, r.ID)
		if len(matches) > 0 {
			fmt.Printf("  (Includes all grants from: %s)\n", strings.Join(matches, ", "))
		}

		for _, g := range grants {
			extras := ""
			if g.Item.Valid {
				extras += fmt.Sprintf(" Item=%s", g.Item.String)
			}
			if g.ItemID.Valid {
				extras += fmt.Sprintf(" ItemID=%d", g.ItemID.Int32)
			}
			if g.Extra.Valid {
				extras += fmt.Sprintf(" Extra=%s", g.Extra.String)
			}
			fmt.Printf("  - Section: %s, Action: %s%s\n", g.Section, g.Action, extras)
		}
	} else {
		fmt.Println("Grants: None")
	}

	return nil
}

func checkRoleInclusion(q *db.Queries, ctx context.Context, targetGrants []*db.Grant, targetRoleID int32) []string {
	allRoles, err := q.AdminListRoles(ctx)
	if err != nil {
		return nil
	}

	var matches []string

	// Convert target grants to a set of signatures for easy lookup
	targetGrantSet := make(map[string]bool)
	for _, g := range targetGrants {
		targetGrantSet[grantSignature(g)] = true
	}

	for _, otherRole := range allRoles {
		if otherRole.ID == targetRoleID {
			continue
		}

		otherGrants, err := q.AdminListGrantsByRoleID(ctx, sql.NullInt32{Int32: otherRole.ID, Valid: true})
		if err != nil {
			continue
		}

		if len(otherGrants) == 0 {
			continue
		}

		// Check if all grants of otherRole are present in targetRole
		allFound := true
		for _, og := range otherGrants {
			if !targetGrantSet[grantSignature(og)] {
				allFound = false
				break
			}
		}

		if allFound {
			matches = append(matches, otherRole.Name)
		}
	}
	return matches
}

func grantSignature(g *db.Grant) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%v",
		g.Section,
		g.Item.String,
		g.RuleType,
		g.ItemRule.String,
		g.ItemID.Int32,
		g.Action,
		g.Extra.String,
		g.Active,
	)
}

func (c *roleInspectCmd) Usage() {
	executeUsage(c.fs.Output(), "role_inspect_usage.txt", c)
}

func (c *roleInspectCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleInspectCmd)(nil)
