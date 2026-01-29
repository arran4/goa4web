package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/permissions"
)

type grantListAvailableCmd struct {
	fs           *flag.FlagSet
	asCliCommand bool
	asJSON       bool
}

func newGrantListAvailableCmd() *grantListAvailableCmd {
	c := &grantListAvailableCmd{
		fs: flag.NewFlagSet("list-available", flag.ContinueOnError),
	}
	c.fs.SetOutput(os.Stdout)
	c.fs.BoolVar(&c.asCliCommand, "as-cli-command", false, "Output as CLI commands instead of a table.")
	c.fs.BoolVar(&c.asJSON, "json", false, "Output grants as JSON grouped by section and item.")
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
		return writeGrantDefinitionsAsCLI(c.fs.Output(), permissions.Definitions)
	}
	if c.asJSON {
		return writeGrantDefinitionsAsJSON(c.fs.Output(), permissions.Definitions)
	}

	return writeGrantDefinitionsAsTable(c.fs.Output(), permissions.Definitions)
}

type grantExportCmd struct {
	fs           *flag.FlagSet
	asCliCommand bool
	asJSON       bool
}

func newGrantExportCmd() *grantExportCmd {
	c := &grantExportCmd{
		fs: flag.NewFlagSet("export", flag.ContinueOnError),
	}
	c.fs.SetOutput(os.Stdout)
	c.fs.BoolVar(&c.asCliCommand, "as-cli-command", false, "Output as CLI commands instead of JSON.")
	c.fs.BoolVar(&c.asJSON, "json", false, "Output grants as JSON grouped by section and item.")
	return c
}

func (c *grantExportCmd) Name() string {
	return c.fs.Name()
}

func (c *grantExportCmd) Init(args []string) error {
	return c.fs.Parse(args)
}

func (c *grantExportCmd) Run() error {
	if c.asCliCommand {
		return writeGrantDefinitionsAsCLI(c.fs.Output(), permissions.Definitions)
	}
	if !c.asJSON {
		c.asJSON = true
	}
	return writeGrantDefinitionsAsJSON(c.fs.Output(), permissions.Definitions)
}

type grantExportAction struct {
	Action      string `json:"action"`
	Description string `json:"description"`
}

type grantExportItem struct {
	Item    string              `json:"item"`
	Actions []grantExportAction `json:"actions"`
}

type grantExportSection struct {
	Section string            `json:"section"`
	Items   []grantExportItem `json:"items"`
}

type grantExportPayload struct {
	Sections []grantExportSection `json:"sections"`
}

func writeGrantDefinitionsAsTable(w io.Writer, defs []*permissions.GrantDefinition) error {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tw, "Section\tItem\tAction\tDescription\tExample")
	for _, def := range defs {
		example := fmt.Sprintf("goa4web grant add -section %s -item %s -action %s <role or user details>", def.Section, def.Item, def.Action)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", def.Section, def.Item, def.Action, def.Description, example)
	}
	return tw.Flush()
}

func writeGrantDefinitionsAsCLI(w io.Writer, defs []*permissions.GrantDefinition) error {
	if w == nil {
		w = os.Stdout
	}
	for _, def := range defs {
		fmt.Fprintf(w, "goa4web grant add -section %s -item %s -action %s <-user-[id] / -role[-id]>\n", def.Section, def.Item, def.Action)
	}
	return nil
}

func writeGrantDefinitionsAsJSON(w io.Writer, defs []*permissions.GrantDefinition) error {
	if w == nil {
		w = os.Stdout
	}
	payload := grantExportPayload{Sections: buildGrantExportSections(defs)}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func buildGrantExportSections(defs []*permissions.GrantDefinition) []grantExportSection {
	var sections []grantExportSection
	sectionIndex := map[string]int{}
	itemIndex := map[string]map[string]int{}
	for _, def := range defs {
		sIndex, ok := sectionIndex[def.Section]
		if !ok {
			sections = append(sections, grantExportSection{Section: def.Section})
			sIndex = len(sections) - 1
			sectionIndex[def.Section] = sIndex
		}
		if itemIndex[def.Section] == nil {
			itemIndex[def.Section] = map[string]int{}
		}
		items := itemIndex[def.Section]
		iIndex, ok := items[def.Item]
		if !ok {
			sections[sIndex].Items = append(sections[sIndex].Items, grantExportItem{Item: def.Item})
			iIndex = len(sections[sIndex].Items) - 1
			items[def.Item] = iIndex
		}
		sections[sIndex].Items[iIndex].Actions = append(sections[sIndex].Items[iIndex].Actions, grantExportAction{
			Action:      def.Action,
			Description: def.Description,
		})
	}
	return sections
}
