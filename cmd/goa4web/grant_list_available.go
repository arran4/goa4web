package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/permissions"
)

type grantListAvailableCmd struct {
	fs           *flag.FlagSet
	asCliCommand bool
}

func newGrantListAvailableCmd() *grantListAvailableCmd {
	c := &grantListAvailableCmd{
		fs: flag.NewFlagSet("list-available", flag.ContinueOnError),
	}
	c.fs.BoolVar(&c.asCliCommand, "as-cli-command", false, "Output as CLI commands instead of a table.")
	return c
}

func (c *grantListAvailableCmd) Name() string {
	return c.fs.Name()
}

func (c *grantListAvailableCmd) Init(args []string) error {
	return c.fs.Parse(args)
}

func (c *grantListAvailableCmd) Run() error {
	if c.asCliCommand {
		for _, def := range permissions.Definitions {
			fmt.Printf("goa4web grant add -section %s -item %s -action %s <-user-[id] / -role[-id]>\n", def.Section, def.Item, def.Action)
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Section\tItem\tAction\tDescription\tExample")
	for _, def := range permissions.Definitions {
		example := fmt.Sprintf("goa4web grant add -section %s -item %s -action %s <role or user details>", def.Section, def.Item, def.Action)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", def.Section, def.Item, def.Action, def.Description, example)
	}
	return w.Flush()
}
