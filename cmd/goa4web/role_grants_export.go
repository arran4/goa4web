package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/internal/db"
)

const (
	// roleGrantsExportFormatJSON emits the canonical JSON schema.
	roleGrantsExportFormatJSON = "json"
	// roleGrantsExportFormatCSV emits CSV rows with JSON-encoded fields.
	roleGrantsExportFormatCSV = "csv"
)

// roleGrantsCmd implements the "role grants" subcommand.
type roleGrantsCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	args []string
}

func parseRoleGrantsCmd(parent *roleCmd, args []string) (*roleGrantsCmd, error) {
	c := &roleGrantsCmd{roleCmd: parent}
	fs := flag.NewFlagSet("grants", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *roleGrantsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing grants subcommand")
	}
	if err := usageIfHelp(c.fs, c.args); err != nil {
		return err
	}
	switch c.args[0] {
	case "export":
		cmd, err := parseRoleGrantsExportCmd(c, c.args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown grants subcommand %q", c.args[0])
	}
}

func (c *roleGrantsCmd) Usage() { executeUsage(c.fs.Output(), "role_grants_usage.txt", c) }

func (c *roleGrantsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleGrantsCmd)(nil)

// roleGrantsExportCmd implements "role grants export".
type roleGrantsExportCmd struct {
	*roleGrantsCmd
	fs     *flag.FlagSet
	role   string
	format string
}

func parseRoleGrantsExportCmd(parent *roleGrantsCmd, args []string) (*roleGrantsExportCmd, error) {
	c := &roleGrantsExportCmd{roleGrantsCmd: parent}
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.StringVar(&c.role, "role", "", "The role name to export.")
	fs.StringVar(&c.format, "format", roleGrantsExportFormatJSON, "Output format: json or csv.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	c.format = strings.ToLower(strings.TrimSpace(c.format))
	if c.format != roleGrantsExportFormatJSON && c.format != roleGrantsExportFormatCSV {
		return nil, fmt.Errorf("unsupported format %q", c.format)
	}
	return c, nil
}

func (c *roleGrantsExportCmd) Run() error {
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("queries: %w", err)
	}
	export, err := buildRoleGrantsExport(c.rootCmd.Context(), queries, c.role, c.rootCmd.cfg)
	if err != nil {
		return err
	}
	switch c.format {
	case roleGrantsExportFormatJSON:
		return writeRoleGrantsExportJSON(c.fs.Output(), export)
	case roleGrantsExportFormatCSV:
		return writeRoleGrantsExportCSV(c.fs.Output(), export)
	default:
		return fmt.Errorf("unsupported format %q", c.format)
	}
}

func (c *roleGrantsExportCmd) Usage() { executeUsage(c.fs.Output(), "role_grants_export_usage.txt", c) }

func (c *roleGrantsExportCmd) FlagGroups() []flagGroup {
	return append(c.roleCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleGrantsExportCmd)(nil)

type roleGrantsExport struct {
	Role        roleGrantsExportRole    `json:"role"`
	GrantGroups []roleGrantsExportGroup `json:"grant_groups"`
}

type roleGrantsExportRole struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	CanLogin      bool   `json:"can_login"`
	IsAdmin       bool   `json:"is_admin"`
	PrivateLabels bool   `json:"private_labels"`
}

type roleGrantsExportItemID struct {
	Int32 int32 `json:"int32"`
	Valid bool  `json:"valid"`
}

type roleGrantsExportAction struct {
	Name        string `json:"name"`
	Unsupported bool   `json:"unsupported"`
}

type roleGrantsExportGroup struct {
	Section     string                   `json:"section"`
	Item        string                   `json:"item"`
	ItemID      roleGrantsExportItemID   `json:"item_id"`
	Link        string                   `json:"link"`
	Info        string                   `json:"info"`
	Have        []roleGrantsExportAction `json:"have"`
	Disabled    []roleGrantsExportAction `json:"disabled"`
	Available   []string                 `json:"available"`
	Unsupported bool                     `json:"unsupported"`
}

func buildRoleGrantsExport(ctx context.Context, queries db.Querier, roleName string, cfg *config.RuntimeConfig) (roleGrantsExport, error) {
	role, err := queries.GetRoleByName(ctx, roleName)
	if err != nil {
		return roleGrantsExport{}, fmt.Errorf("get role %q: %w", roleName, err)
	}
	if role == nil {
		return roleGrantsExport{}, fmt.Errorf("get role %q: %w", roleName, sql.ErrNoRows)
	}

	cd := common.NewCoreData(ctx, queries, cfg)
	groups, err := adminhandlers.BuildGrantGroups(ctx, cd, role.ID)
	if err != nil {
		return roleGrantsExport{}, fmt.Errorf("build grant groups: %w", err)
	}

	out := roleGrantsExport{
		Role: roleGrantsExportRole{
			ID:            role.ID,
			Name:          role.Name,
			CanLogin:      role.CanLogin,
			IsAdmin:       role.IsAdmin,
			PrivateLabels: role.PrivateLabels,
		},
		GrantGroups: make([]roleGrantsExportGroup, 0, len(groups)),
	}

	for _, group := range groups {
		out.GrantGroups = append(out.GrantGroups, roleGrantsExportGroup{
			Section:     group.Section,
			Item:        group.Item,
			ItemID:      roleGrantsExportItemID{Int32: group.ItemID.Int32, Valid: group.ItemID.Valid},
			Link:        group.Link,
			Info:        group.Info,
			Have:        convertRoleGrantsExportActions(group.Have),
			Disabled:    convertRoleGrantsExportActions(group.Disabled),
			Available:   group.Available,
			Unsupported: group.Unsupported,
		})
	}

	return out, nil
}

func convertRoleGrantsExportActions(actions []adminhandlers.GrantAction) []roleGrantsExportAction {
	out := make([]roleGrantsExportAction, 0, len(actions))
	for _, action := range actions {
		out = append(out, roleGrantsExportAction{
			Name:        action.Name,
			Unsupported: action.Unsupported,
		})
	}
	return out
}

func writeRoleGrantsExportJSON(w io.Writer, export roleGrantsExport) error {
	payload, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if _, err := fmt.Fprintln(w, string(payload)); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

func writeRoleGrantsExportCSV(w io.Writer, export roleGrantsExport) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{
		"role_id",
		"role_name",
		"role_can_login",
		"role_is_admin",
		"role_private_labels",
		"section",
		"item",
		"item_id",
		"info",
		"link",
		"have",
		"disabled",
		"available",
		"unsupported",
	}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, group := range export.GrantGroups {
		itemIDJSON, err := json.Marshal(group.ItemID)
		if err != nil {
			return fmt.Errorf("marshal item id: %w", err)
		}
		haveJSON, err := json.Marshal(group.Have)
		if err != nil {
			return fmt.Errorf("marshal have actions: %w", err)
		}
		disabledJSON, err := json.Marshal(group.Disabled)
		if err != nil {
			return fmt.Errorf("marshal disabled actions: %w", err)
		}
		availableJSON, err := json.Marshal(group.Available)
		if err != nil {
			return fmt.Errorf("marshal available actions: %w", err)
		}
		row := []string{
			fmt.Sprintf("%d", export.Role.ID),
			export.Role.Name,
			fmt.Sprintf("%t", export.Role.CanLogin),
			fmt.Sprintf("%t", export.Role.IsAdmin),
			fmt.Sprintf("%t", export.Role.PrivateLabels),
			group.Section,
			group.Item,
			string(itemIDJSON),
			group.Info,
			group.Link,
			string(haveJSON),
			string(disabledJSON),
			string(availableJSON),
			fmt.Sprintf("%t", group.Unsupported),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}
