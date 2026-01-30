package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
)

// dlqDeleteCmd implements "dlq delete".
type dlqDeleteCmd struct {
	*dlqCmd
	fs      *flag.FlagSet
	id      int
	ids     string
	jsonOut bool
}

func parseDlqDeleteCmd(parent *dlqCmd, args []string) (*dlqDeleteCmd, error) {
	c := &dlqDeleteCmd{dlqCmd: parent}
	fs, _, err := parseFlags("delete", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.id, "id", 0, "dead letter ID to delete")
		fs.StringVar(&c.ids, "ids", "", "comma-separated list of dead letter IDs to delete")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

func (c *dlqDeleteCmd) Run() error {
	providers, err := c.providers()
	if err != nil {
		return err
	}
	if !dlqHasProvider(providers, "db") {
		return fmt.Errorf("db dlq provider not configured")
	}
	ids, err := c.parseIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("id required")
	}
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	for _, id := range ids {
		if err := queries.SystemDeleteDeadLetter(c.rootCmd.Context(), id); err != nil {
			return fmt.Errorf("delete dead letter %d: %w", id, err)
		}
	}
	if c.jsonOut {
		out := map[string]interface{}{
			"deleted_ids": ids,
			"count":       len(ids),
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}
	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tStatus")
	for _, id := range ids {
		fmt.Fprintf(w, "%d\tdeleted\n", id)
	}
	return w.Flush()
}

func (c *dlqDeleteCmd) parseIDs() ([]int32, error) {
	ids := make([]int32, 0)
	if c.id != 0 {
		ids = append(ids, int32(c.id))
	}
	if c.ids == "" {
		return ids, nil
	}
	parts := strings.Split(c.ids, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse id %q: %w", part, err)
		}
		ids = append(ids, int32(val))
	}
	return ids, nil
}

// Usage prints command usage information with examples.
func (c *dlqDeleteCmd) Usage() {
	executeUsage(c.fs.Output(), "dlq_delete_usage.txt", c)
}

func (c *dlqDeleteCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dlqDeleteCmd)(nil)
