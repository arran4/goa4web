package main

import (
	"context"
	"flag"
	"fmt"
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSection\tItem\tAction\tRule\tRole\tUser\tActive")

	for _, g := range rows {
		item := g.Item.String
		if !g.Item.Valid {
			item = "-"
		}

		role := "-"
		if g.RoleName.Valid {
			role = g.RoleName.String
		} else if g.RoleID.Valid {
			role = fmt.Sprintf("ID:%d", g.RoleID.Int32)
		}

		user := "-"
		if g.Username.Valid {
			user = g.Username.String
		} else if g.UserID.Valid {
			user = fmt.Sprintf("ID:%d", g.UserID.Int32)
		}

		active := "No"
		if g.Active {
			active = "Yes"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			g.ID,
			g.Section,
			item,
			g.Action,
			g.RuleType,
			role,
			user,
			active,
		)
	}
	return w.Flush()
}
