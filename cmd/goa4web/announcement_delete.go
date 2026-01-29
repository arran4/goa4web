package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// announcementDeleteCmd implements "announcement delete".
type announcementDeleteCmd struct {
	*announcementCmd
	fs      *flag.FlagSet
	id      int
	ids     string
	jsonOut bool
}

func parseAnnouncementDeleteCmd(parent *announcementCmd, args []string) (*announcementDeleteCmd, error) {
	c := &announcementDeleteCmd{announcementCmd: parent}
	fs, _, err := parseFlags("delete", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.id, "id", 0, "announcement ID to delete")
		fs.StringVar(&c.ids, "ids", "", "comma-separated list of announcement IDs to delete")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

// Usage prints command usage information with examples.
func (c *announcementDeleteCmd) Usage() {
	executeUsage(c.fs.Output(), "announcement_delete_usage.txt", c)
}

func (c *announcementDeleteCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*announcementDeleteCmd)(nil)

func (c *announcementDeleteCmd) Run() error {
	ids, err := c.parseIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	for _, id := range ids {
		if err := queries.AdminDemoteAnnouncement(ctx, id); err != nil {
			return fmt.Errorf("demote announcement: %w", err)
		}
	}
	if c.jsonOut {
		out := map[string]interface{}{
			"deleted_ids": ids,
			"count":       len(ids),
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tStatus")
	for _, id := range ids {
		fmt.Fprintf(w, "%d\tdeleted\n", id)
	}
	return w.Flush()
}

func (c *announcementDeleteCmd) parseIDs() ([]int32, error) {
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
