package main

import (
	"flag"
	"fmt"
	"strings"
)

// dlqCmd handles dead letter queue commands.
type dlqCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseDlqCmd(parent *rootCmd, args []string) (*dlqCmd, error) {
	c := &dlqCmd{rootCmd: parent}
	c.fs = newFlagSet("dlq")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *dlqCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing dlq command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseDlqListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseDlqDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "purge":
		cmd, err := parseDlqPurgeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("purge: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown dlq command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *dlqCmd) Usage() {
	executeUsage(c.fs.Output(), "dlq_usage.txt", c)
}

func (c *dlqCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

func (c *dlqCmd) providers() ([]string, error) {
	cfg := c.rootCmd.cfg
	if cfg == nil {
		return nil, fmt.Errorf("runtime config not initialized")
	}
	if strings.TrimSpace(cfg.DLQProvider) == "" {
		return nil, fmt.Errorf("dlq provider not configured")
	}
	names := strings.Split(cfg.DLQProvider, ",")
	providers := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		providers = append(providers, name)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("dlq provider not configured")
	}
	return providers, nil
}

func dlqHasProvider(providers []string, target string) bool {
	for _, provider := range providers {
		if provider == target {
			return true
		}
	}
	return false
}

var _ usageData = (*dlqCmd)(nil)
