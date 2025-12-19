package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// roleTemplateCmd implements the "role template" subcommand.
type roleTemplateCmd struct {
	*roleCmd
	fs *flag.FlagSet
}

func parseRoleTemplateCmd(parent *roleCmd, args []string) (*roleTemplateCmd, error) {
	c := &roleTemplateCmd{roleCmd: parent}
	fs := flag.NewFlagSet("template", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleTemplateCmd) Run() error {
	if c.fs.NArg() == 0 {
		c.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch c.fs.Arg(0) {
	case "list":
		cmd, err := parseRoleTemplateListCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "explain":
		cmd, err := parseRoleTemplateExplainCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "setup":
		cmd, err := parseRoleTemplateSetupCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "diff":
		cmd, err := parseRoleTemplateDiffCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.Usage()
		return fmt.Errorf("unknown subcommand: %s", c.fs.Arg(0))
	}
}

func (c *roleTemplateCmd) Usage() {
	executeUsage(c.fs.Output(), "role_template_usage.txt", c)
}

func (c *roleTemplateCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleTemplateCmd)(nil)

// roleTemplateListCmd implements "role template list"
type roleTemplateListCmd struct {
	*roleTemplateCmd
	fs *flag.FlagSet
}

func parseRoleTemplateListCmd(parent *roleTemplateCmd, args []string) (*roleTemplateListCmd, error) {
	c := &roleTemplateListCmd{roleTemplateCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleTemplateListCmd) Run() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tDescription")

	var names []string
	for k := range roleScenarios {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		sc := roleScenarios[name]
		fmt.Fprintf(w, "%s\t%s\n", sc.Name, sc.Description)
	}
	w.Flush()
	return nil
}

func (c *roleTemplateListCmd) Usage() {
	executeUsage(c.fs.Output(), "role_template_list_usage.txt", c)
}

func (c *roleTemplateListCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleTemplateListCmd)(nil)

// roleTemplateExplainCmd implements "role template explain"
type roleTemplateExplainCmd struct {
	*roleTemplateCmd
	fs   *flag.FlagSet
	name string
}

func parseRoleTemplateExplainCmd(parent *roleTemplateCmd, args []string) (*roleTemplateExplainCmd, error) {
	c := &roleTemplateExplainCmd{roleTemplateCmd: parent}
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.name, "name", "", "The template name to explain.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	return c, nil
}

func (c *roleTemplateExplainCmd) Run() error {
	sc, ok := roleScenarios[c.name]
	if !ok {
		return fmt.Errorf("template %q not found", c.name)
	}

	fmt.Printf("Template: %s\n", sc.Name)
	fmt.Printf("Description: %s\n\n", sc.Description)
	fmt.Println("Roles Defined:")

	for _, r := range sc.Roles {
		fmt.Printf("  - %s: %s\n", r.Name, r.Description)
		fmt.Printf("    Login: %v, Admin: %v\n", r.CanLogin, r.IsAdmin)
		if len(r.Grants) > 0 {
			fmt.Println("    Grants:")
			for _, g := range r.Grants {
				item := g.Item
				if item == "" { item = "*" }
				fmt.Printf("      - %s / %s / %s (ID: %d)\n", g.Section, item, g.Action, g.ItemID)
			}
		} else {
			fmt.Println("    (No Grants)")
		}
		fmt.Println()
	}
	return nil
}

func (c *roleTemplateExplainCmd) Usage() {
	executeUsage(c.fs.Output(), "role_template_explain_usage.txt", c)
}

func (c *roleTemplateExplainCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleTemplateExplainCmd)(nil)

// roleTemplateSetupCmd implements "role template setup"
// This reuses the logic from roleSetupCmd but adapted to the template structure.
type roleTemplateSetupCmd struct {
	*roleTemplateCmd
	fs   *flag.FlagSet
	name string
}

func parseRoleTemplateSetupCmd(parent *roleTemplateCmd, args []string) (*roleTemplateSetupCmd, error) {
	c := &roleTemplateSetupCmd{roleTemplateCmd: parent}
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.name, "name", "", "The template name to apply.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	return c, nil
}

func (c *roleTemplateSetupCmd) Run() error {
	// Reusing the logic from roleSetupCmd is tricky because roleSetupCmd is a struct.
	// But we can extract the logic or just instantiate a helper.
	// However, roleSetupCmd expects `roleCmd` parent, here we have `roleTemplateCmd`.
	// Let's create a helper function `applyScenario` that takes db and scenario def.

	sc, ok := roleScenarios[c.name]
	if !ok {
		return fmt.Errorf("template %q not found", c.name)
	}

	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	ctx := c.rootCmd.ctx
	tx, err := sdb.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	q := db.New(tx)

	fmt.Println("--- BEFORE STATE ---")
	if err := printRolesState(ctx, q, sc.Roles); err != nil {
		return err
	}

	fmt.Printf("\nApplying template %q...\n", sc.Name)
	if err := applyRoles(ctx, q, tx, sc.Roles); err != nil {
		return err
	}

	fmt.Println("\n--- AFTER STATE ---")
	if err := printRolesState(ctx, q, sc.Roles); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (c *roleTemplateSetupCmd) Usage() {
	executeUsage(c.fs.Output(), "role_template_setup_usage.txt", c)
}

func (c *roleTemplateSetupCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleTemplateSetupCmd)(nil)

// roleTemplateDiffCmd implements "role template diff"
type roleTemplateDiffCmd struct {
	*roleTemplateCmd
	fs   *flag.FlagSet
	name string
}

func parseRoleTemplateDiffCmd(parent *roleTemplateCmd, args []string) (*roleTemplateDiffCmd, error) {
	c := &roleTemplateDiffCmd{roleTemplateCmd: parent}
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.name, "name", "", "The template name to diff.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	return c, nil
}

func (c *roleTemplateDiffCmd) Run() error {
	sc, ok := roleScenarios[c.name]
	if !ok {
		return fmt.Errorf("template %q not found", c.name)
	}

	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	fmt.Printf("Diff for template %q:\n", sc.Name)

	for _, rDef := range sc.Roles {
		fmt.Printf("\nRole: %s\n", rDef.Name)
		role, err := q.GetRoleByName(ctx, rDef.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("  Status: New (Will be created)")
				continue
			}
			return fmt.Errorf("get role %s: %w", rDef.Name, err)
		}

		// Role Properties Diff
		changes := []string{}
		if role.CanLogin != rDef.CanLogin { changes = append(changes, fmt.Sprintf("CanLogin: %v -> %v", role.CanLogin, rDef.CanLogin)) }
		if role.IsAdmin != rDef.IsAdmin { changes = append(changes, fmt.Sprintf("IsAdmin: %v -> %v", role.IsAdmin, rDef.IsAdmin)) }

		if len(changes) > 0 {
			fmt.Println("  Properties Update:")
			for _, ch := range changes {
				fmt.Printf("    - %s\n", ch)
			}
		} else {
			fmt.Println("  Properties: Match")
		}

		// Grants Diff
		// Since setup wipes and recreates, effectively all current grants are removed and new ones added.
		// But a smart diff would show what is actually different.
		currentGrants, err := q.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: role.ID, Valid: true})
		if err != nil {
			return err
		}

		// Simple logic:
		// 1. Convert current to set of strings
		// 2. Convert desired to set of strings
		// 3. Compare

		currSet := make(map[string]bool)
		for _, g := range currentGrants {
			key := grantKey(g.Section, g.Item.String, g.Action, g.ItemID.Int32)
			currSet[key] = true
		}

		fmt.Println("  Grants:")
		hasDiff := false
		for _, g := range rDef.Grants {
			key := grantKey(g.Section, g.Item, g.Action, g.ItemID)
			if !currSet[key] {
				fmt.Printf("    + %s\n", key)
				hasDiff = true
			} else {
				// Mark as seen so we can find deletions
				currSet[key] = false
			}
		}

		for key, exists := range currSet {
			if exists {
				fmt.Printf("    - %s\n", key)
				hasDiff = true
			}
		}

		if !hasDiff {
			fmt.Println("    (No Changes)")
		}
	}

	return nil
}

func grantKey(section, item, action string, itemID int32) string {
	if item == "" { item = "*" }
	return fmt.Sprintf("%s / %s / %s [%d]", section, item, action, itemID)
}

func (c *roleTemplateDiffCmd) Usage() {
	executeUsage(c.fs.Output(), "role_template_diff_usage.txt", c)
}

func (c *roleTemplateDiffCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleTemplateDiffCmd)(nil)
