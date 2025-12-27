package main

import (
	"flag"
	"fmt"
)

// faqCmd handles FAQ management subcommands.
type faqCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseFaqCmd(parent *rootCmd, args []string) (*faqCmd, error) {
	c := &faqCmd{rootCmd: parent}
	c.fs = newFlagSet("faq")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing faq command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "tree":
		cmd, err := parseFaqTreeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseFaqReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "add-from-template":
		cmd, err := parseFaqAddFromTemplateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add-from-template: %w", err)
		}
		return cmd.Run()
	case "list-templates":
		cmd, err := parseFaqListTemplatesCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-templates: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseFaqListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "dump":
		cmd, err := parseFaqDumpCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseFaqUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "reorder":
		cmd, err := parseFaqReorderCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("reorder: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseFaqDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "create":
		cmd, err := parseFaqCreateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown faq command %q", args[0])
	}
}

func (c *faqCmd) Usage() {
	executeUsage(c.fs.Output(), "faq_usage.txt", c)
}

func (c *faqCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCmd)(nil)
