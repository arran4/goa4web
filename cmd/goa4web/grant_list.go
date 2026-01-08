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
}

func parseGrantListCmd(parent *grantCmd, args []string) (*grantListCmd, error) {
	c := &grantListCmd{grantCmd: parent}
	c.fs = newFlagSet("list")
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
	rows, err := q.ListGrantsExtended(ctx)
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
	fmt.Fprintln(w, "ID\tSection\tItem\tAction\tRule Type\tTarget")
	for _, g := range rows {
		target := ""
		if g.RoleName.Valid {
			target = fmt.Sprintf("Role: %s", g.RoleName.String)
		} else if g.Username.Valid {
			target = fmt.Sprintf("User: %s", g.Username.String)
		} else if g.RoleID.Valid {
			target = fmt.Sprintf("Role ID: %d", g.RoleID.Int32)
		} else if g.UserID.Valid {
			target = fmt.Sprintf("User ID: %d", g.UserID.Int32)
		}
		item := ""
		if g.Item.Valid {
			item = g.Item.String
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", g.ID, g.Section, item, g.Action, g.RuleType, target)
	}
	return w.Flush()
}
