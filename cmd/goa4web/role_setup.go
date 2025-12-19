package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// roleSetupCmd implements the "role setup" subcommand.
type roleSetupCmd struct {
	*roleCmd
	fs       *flag.FlagSet
	scenario string
}

func parseRoleSetupCmd(parent *roleCmd, args []string) (*roleSetupCmd, error) {
	c := &roleSetupCmd{roleCmd: parent}
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.scenario, "scenario", "default", "The scenario to apply (default: news/forum setup)")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

type RoleDef struct {
	Name        string
	CanLogin    bool
	IsAdmin     bool
	Description string
	Grants      []GrantDef
}

type GrantDef struct {
	Section string
	Item    string // can be empty
	Action  string
	ItemID  int32 // 0 for global/any
}

func (c *roleSetupCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	ctx := c.rootCmd.ctx

	// Start Transaction
	tx, err := sdb.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	q := db.New(tx)

	// Define scenarios
	var rolesToSetup []RoleDef

	switch c.scenario {
	case "default":
		rolesToSetup = []RoleDef{
			{
				Name:     "guest",
				CanLogin: false,
				IsAdmin:  false,
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
				},
			},
			{
				Name:     "user",
				CanLogin: true,
				IsAdmin:  false,
				Grants: []GrantDef{
					// News reader
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					// Labeller
					{Section: "news", Item: "post", Action: "label"},
					{Section: "privateforum", Item: "topic", Action: "label"},
					// Private forum user
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
				},
			},
			{
				Name:     "admin",
				CanLogin: true,
				IsAdmin:  true,
				Grants: []GrantDef{
					// News writer
					{Section: "news", Item: "post", Action: "post"},
					{Section: "news", Item: "post", Action: "edit"},
					{Section: "news", Item: "post", Action: "reply"},
					// News reader
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					// Labeller
					{Section: "news", Item: "post", Action: "label"},
					{Section: "privateforum", Item: "topic", Action: "label"},
					// Private forum user
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
					{Section: "privateforum", Item: "topic", Action: "post"},
					{Section: "privateforum", Item: "topic", Action: "reply"},
					{Section: "privateforum", Item: "topic", Action: "edit"},
				},
			},
		}
	default:
		return fmt.Errorf("unknown scenario: %s", c.scenario)
	}

	// 1. Report "Before" state (Using same transaction/snapshot)
	fmt.Println("--- BEFORE STATE ---")
	if err := c.printRolesState(ctx, q, rolesToSetup); err != nil {
		return err
	}

	// 2. Apply Changes
	fmt.Println("\nApplying changes...")
	for _, rDef := range rolesToSetup {
		// Create or Update Role
		role, err := q.GetRoleByName(ctx, rDef.Name)
		var roleID int32
		if err != nil {
			if err == sql.ErrNoRows {
				// Create
				res, err := tx.ExecContext(ctx, "INSERT INTO roles (name, can_login, is_admin, private_labels, public_profile_allowed_at) VALUES (?, ?, ?, ?, NOW())", rDef.Name, rDef.CanLogin, rDef.IsAdmin, rDef.CanLogin)
				if err != nil {
					return fmt.Errorf("create role %s: %w", rDef.Name, err)
				}
				id, err := res.LastInsertId()
				if err != nil {
					return fmt.Errorf("get last insert id: %w", err)
				}
				roleID = int32(id)
				log.Printf("Created role %s (ID: %d)", rDef.Name, roleID)
			} else {
				return fmt.Errorf("get role %s: %w", rDef.Name, err)
			}
		} else {
			roleID = role.ID
			// Update existing role properties
			// AdminUpdateRole uses queries-roles.sql which is available
			if err := q.AdminUpdateRole(ctx, db.AdminUpdateRoleParams{
				Name:          rDef.Name,
				CanLogin:      rDef.CanLogin,
				IsAdmin:       rDef.IsAdmin,
				PrivateLabels: rDef.CanLogin,
				ID:            role.ID,
			}); err != nil {
				return fmt.Errorf("update role %s: %w", rDef.Name, err)
			}
			log.Printf("Updated role %s (ID: %d)", rDef.Name, roleID)
		}

		// Delete existing grants
		if err := q.DeleteGrantsByRoleID(ctx, sql.NullInt32{Int32: roleID, Valid: true}); err != nil {
			return fmt.Errorf("delete grants for role %s: %w", rDef.Name, err)
		}

		// Insert new grants
		for _, g := range rDef.Grants {
			err := q.CreateGrant(ctx, db.CreateGrantParams{
				RoleID:   sql.NullInt32{Int32: roleID, Valid: true},
				Section:  g.Section,
				Item:     sql.NullString{String: g.Item, Valid: g.Item != ""},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
				Action:   g.Action,
				Active:   true,
			})
			if err != nil {
				return fmt.Errorf("create grant for %s (%s/%s/%s): %w", rDef.Name, g.Section, g.Item, g.Action, err)
			}
		}
	}

	// 3. Report "After" state
	fmt.Println("\n--- AFTER STATE ---")
	if err := c.printRolesState(ctx, q, rolesToSetup); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (c *roleSetupCmd) printRolesState(ctx context.Context, q *db.Queries, roles []RoleDef) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// Improved headers
	fmt.Fprintln(w, "Role\tLogin\tAdmin\tSection\tItem\tAction\tItemID")

	for _, rDef := range roles {
		role, err := q.GetRoleByName(ctx, rDef.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Fprintf(w, "%s\t-\t-\t(Not Found)\t\t\t\n", rDef.Name)
				continue
			}
			return err
		}

		grants, err := q.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: role.ID, Valid: true})
		if err != nil {
			return err
		}

		roleInfo := fmt.Sprintf("%s\t%v\t%v", role.Name, role.CanLogin, role.IsAdmin)

		if len(grants) == 0 {
			fmt.Fprintf(w, "%s\t(No Grants)\t\t\t\n", roleInfo)
		}

		for _, g := range grants {
			item := g.Item.String
			if !g.Item.Valid { item = "*" }
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", roleInfo, g.Section, item, g.Action, g.ItemID.Int32)
		}
	}
	w.Flush()
	return nil
}

func (c *roleSetupCmd) Usage() {
	executeUsage(c.fs.Output(), "role_setup_usage.txt", c)
}

func (c *roleSetupCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleSetupCmd)(nil)
