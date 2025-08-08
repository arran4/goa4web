package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// linksRemapApplyCmd implements "links remap apply".
type linksRemapApplyCmd struct {
	*linksRemapCmd
	fs   *flag.FlagSet
	File string
}

func parseLinksRemapApplyCmd(parent *linksRemapCmd, args []string) (*linksRemapApplyCmd, error) {
	c := &linksRemapApplyCmd{linksRemapCmd: parent}
	fs, _, err := parseFlags("apply", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.File, "file", "", "CSV file with remappings")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *linksRemapApplyCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	f, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	for i, rec := range records {
		if i == 0 && strings.EqualFold(rec[0], "internal reference") {
			continue
		}
		if len(rec) < 3 || rec[2] == "" {
			continue
		}
		parts := strings.SplitN(rec[0], ":", 2)
		if len(parts) != 2 {
			continue
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		switch parts[0] {
		case "site_news":
			if _, err := conn.ExecContext(ctx, "UPDATE site_news SET news = REPLACE(news, ?, ?) WHERE idsiteNews = ?", rec[1], rec[2], id); err != nil {
				return fmt.Errorf("update news %d: %w", id, err)
			}
			if _, err := conn.ExecContext(ctx, "DELETE FROM external_links WHERE url = ?", rec[1]); err != nil {
				return fmt.Errorf("cleanup external link %q: %w", rec[1], err)
			}
		default:
			continue
		}
	}
	return nil
}

func (c *linksRemapApplyCmd) Usage() {
	executeUsage(c.fs.Output(), "links_remap_apply_usage.txt", c)
}

func (c *linksRemapApplyCmd) FlagGroups() []flagGroup {
	return append(c.linksRemapCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksRemapApplyCmd)(nil)
