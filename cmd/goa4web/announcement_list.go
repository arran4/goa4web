package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// announcementListCmd implements "announcement list".
type announcementListCmd struct {
	*announcementCmd
	fs      *flag.FlagSet
	jsonOut bool
}

func parseAnnouncementListCmd(parent *announcementCmd, args []string) (*announcementListCmd, error) {
	c := &announcementListCmd{announcementCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
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
func (c *announcementListCmd) Usage() {
	executeUsage(c.fs.Output(), "announcement_list_usage.txt", c)
}

func (c *announcementListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*announcementListCmd)(nil)

func (c *announcementListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListAnnouncementsWithNews(ctx)
	if err != nil {
		return fmt.Errorf("list announcements: %w", err)
	}

	if c.jsonOut {
		out := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			var title *string
			if row.News.Valid {
				t := row.News.String
				title = &t
			}
			item := map[string]interface{}{
				"id":         row.ID,
				"news_id":    row.SiteNewsID,
				"active":     row.Active,
				"created_at": row.CreatedAt.Format(time.RFC3339),
				"news_title": title,
			}
			out = append(out, item)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNewsID\tActive\tCreatedAt\tTitle")
	for _, row := range rows {
		fmt.Fprintf(w, "%d\t%d\t%t\t%s\t%s\n", row.ID, row.SiteNewsID, row.Active, row.CreatedAt.Format(time.RFC3339), row.News.String)
	}
	return w.Flush()
}
