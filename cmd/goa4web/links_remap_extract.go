package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/arran4/goa4web/internal/db"
)

// linksRemapExtractCmd implements "links remap extract".
type linksRemapExtractCmd struct {
	*linksRemapCmd
	fs *flag.FlagSet
}

func parseLinksRemapExtractCmd(parent *linksRemapCmd, args []string) (*linksRemapExtractCmd, error) {
	c := &linksRemapExtractCmd{linksRemapCmd: parent}
	fs, _, err := parseFlags("extract", args, nil)
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *linksRemapExtractCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.GetAllSiteNewsForIndex(ctx)
	if err != nil {
		return fmt.Errorf("list news: %w", err)
	}
	w := csv.NewWriter(os.Stdout)
	if err := w.Write([]string{"internal reference", "original url", "to url"}); err != nil {
		return err
	}
	re := regexp.MustCompile(`https?://[^\s"']+`)
	for _, r := range rows {
		if r.News.Valid {
			matches := re.FindAllString(r.News.String, -1)
			for _, m := range matches {
				_ = w.Write([]string{fmt.Sprintf("site_news:%d", r.Idsitenews), m, ""})
			}
		}
	}
	w.Flush()
	return w.Error()
}

func (c *linksRemapExtractCmd) Usage() {
	executeUsage(c.fs.Output(), "links_remap_extract_usage.txt", c)
}

func (c *linksRemapExtractCmd) FlagGroups() []flagGroup {
	return append(c.linksRemapCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksRemapExtractCmd)(nil)
