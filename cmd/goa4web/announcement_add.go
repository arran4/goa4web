package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// announcementAddCmd implements "announcement add".
type announcementAddCmd struct {
	*announcementCmd
	fs      *flag.FlagSet
	newsID  int
	jsonOut bool
}

func parseAnnouncementAddCmd(parent *announcementCmd, args []string) (*announcementAddCmd, error) {
	c := &announcementAddCmd{announcementCmd: parent}
	fs, _, err := parseFlags("add", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.newsID, "news-id", 0, "news ID to promote to announcement")
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
func (c *announcementAddCmd) Usage() {
	executeUsage(c.fs.Output(), "announcement_add_usage.txt", c)
}

func (c *announcementAddCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*announcementAddCmd)(nil)

func (c *announcementAddCmd) Run() error {
	if c.newsID == 0 {
		return fmt.Errorf("news-id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminPromoteAnnouncement(ctx, int32(c.newsID)); err != nil {
		return fmt.Errorf("promote announcement: %w", err)
	}
	if c.jsonOut {
		out := map[string]interface{}{
			"news_id": c.newsID,
			"status":  "added",
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NewsID\tStatus")
	fmt.Fprintf(w, "%d\tadded\n", c.newsID)
	return w.Flush()
}
