package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

// dlqPurgeCmd implements "dlq purge".
type dlqPurgeCmd struct {
	*dlqCmd
	fs      *flag.FlagSet
	before  string
	jsonOut bool
}

func parseDlqPurgeCmd(parent *dlqCmd, args []string) (*dlqPurgeCmd, error) {
	c := &dlqPurgeCmd{dlqCmd: parent}
	fs, _, err := parseFlags("purge", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.before, "before", "", "purge entries before this date (YYYY-MM-DD)")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

func (c *dlqPurgeCmd) Run() error {
	providers, err := c.providers()
	if err != nil {
		return err
	}
	if !dlqHasProvider(providers, "db") {
		return fmt.Errorf("db dlq provider not configured")
	}
	purgeBefore, err := c.parsePurgeBefore()
	if err != nil {
		return err
	}
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := queries.SystemPurgeDeadLettersBefore(c.rootCmd.Context(), purgeBefore); err != nil {
		return fmt.Errorf("purge dead letters: %w", err)
	}
	purgeAt := purgeBefore.Format(time.RFC3339)
	if c.jsonOut {
		out := map[string]interface{}{
			"purged_before": purgeAt,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}
	fmt.Fprintf(c.fs.Output(), "Purged dead letters before %s\n", purgeAt)
	return nil
}

func (c *dlqPurgeCmd) parsePurgeBefore() (time.Time, error) {
	if c.before == "" {
		return time.Now(), nil
	}
	if t, err := time.Parse("2006-01-02", c.before); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, c.before); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid before date %q", c.before)
}

// Usage prints command usage information with examples.
func (c *dlqPurgeCmd) Usage() {
	executeUsage(c.fs.Output(), "dlq_purge_usage.txt", c)
}

func (c *dlqPurgeCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dlqPurgeCmd)(nil)
