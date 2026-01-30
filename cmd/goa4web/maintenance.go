package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// maintenancePrivateForumHandler is the handler name for private forum topics.
	maintenancePrivateForumHandler = "private"
)

// maintenanceCmd handles once-off maintenance tasks.
type maintenanceCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseMaintenanceCmd(parent *rootCmd, args []string) (*maintenanceCmd, error) {
	c := &maintenanceCmd{rootCmd: parent}
	c.fs = newFlagSet("maintenance")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *maintenanceCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing maintenance command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "forum-topic-convert-private":
		cmd, err := parseMaintenanceForumTopicConvertPrivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("forum-topic-convert-private: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown maintenance command %q", args[0])
	}
}

func (c *maintenanceCmd) Usage() {
	executeUsage(c.fs.Output(), "maintenance_usage.txt", c)
}

func (c *maintenanceCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*maintenanceCmd)(nil)

// maintenanceForumTopicConvertPrivateCmd implements "maintenance forum-topic-convert-private".
type maintenanceForumTopicConvertPrivateCmd struct {
	*maintenanceCmd
	fs *flag.FlagSet
}

func parseMaintenanceForumTopicConvertPrivateCmd(parent *maintenanceCmd, args []string) (*maintenanceForumTopicConvertPrivateCmd, error) {
	c := &maintenanceForumTopicConvertPrivateCmd{maintenanceCmd: parent}
	fs, _, err := parseFlags("forum-topic-convert-private", args, nil)
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *maintenanceForumTopicConvertPrivateCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("topic id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	failures := 0
	for _, arg := range args {
		topicID, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Fprintf(c.fs.Output(), "Invalid topic id %q: %v\n", arg, err)
			failures++
			continue
		}
		if err := queries.SystemSetForumTopicHandlerByID(ctx, db.SystemSetForumTopicHandlerByIDParams{
			Handler: maintenancePrivateForumHandler,
			ID:      int32(topicID),
		}); err != nil {
			fmt.Fprintf(c.fs.Output(), "Failed to convert forum topic %d to private: %v\n", topicID, err)
			failures++
			continue
		}
		fmt.Fprintf(c.fs.Output(), "Converted forum topic %d to private.\n", topicID)
	}
	successes := len(args) - failures
	fmt.Fprintf(c.fs.Output(), "Maintenance summary: %d converted, %d failed.\n", successes, failures)
	if failures > 0 {
		return fmt.Errorf("maintenance completed with %d failures", failures)
	}
	return nil
}
