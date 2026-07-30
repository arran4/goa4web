package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type privateForumTopicMergeCmd struct {
	*privateForumCmd
	fs     *flag.FlagSet
	dryRun bool
}

func parsePrivateForumTopicMergeCmd(parent *privateForumCmd, args []string) (*privateForumTopicMergeCmd, error) {
	c := &privateForumTopicMergeCmd{privateForumCmd: parent}
	c.fs = newFlagSet("merge")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *privateForumTopicMergeCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	queries := db.New(conn)
	cd := common.NewCoreData(ctx, queries, nil)

	_, err = cd.MergePrivateTopicsWithSameParticipants(ctx, c.dryRun)
	return err
}
