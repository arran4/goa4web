package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// langDeleteCmd implements "lang delete".
type langDeleteCmd struct {
	*langCmd
	fs      *flag.FlagSet
	ID      int
	Confirm bool
}

func parseLangDeleteCmd(parent *langCmd, args []string) (*langDeleteCmd, error) {
	c := &langDeleteCmd{langCmd: parent}
	c.fs = newFlagSet("delete")
	c.fs.IntVar(&c.ID, "id", 0, "language id")
	c.fs.BoolVar(&c.Confirm, "confirm", false, "confirm deletion")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *langDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	if !c.Confirm {
		return fmt.Errorf("confirm deletion with -confirm")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, nil)
	id, name, err := cd.DeleteLanguage(strconv.Itoa(c.ID))
	if err != nil {
		return fmt.Errorf("delete language: %w", err)
	}
	if name == "" {
		c.rootCmd.Infof("deleted language %d", id)
		return nil
	}
	c.rootCmd.Infof("deleted language %s (%d)", name, id)
	return nil
}

func (c *langDeleteCmd) Usage() {
	executeUsage(c.fs.Output(), "lang_delete_usage.txt", c)
}

func (c *langDeleteCmd) FlagGroups() []flagGroup {
	return append(c.langCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*langDeleteCmd)(nil)
