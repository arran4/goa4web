package main

import (
	"flag"
	"fmt"
)

// helpCmd displays usage information for commands.
type helpCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseHelpCmd(parent *rootCmd, args []string) (*helpCmd, error) {
	c := &helpCmd{rootCmd: parent}
	c.fs = newFlagSet("help")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *helpCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.rootCmd.fs.Usage()
		return nil
	}
	return c.showHelp(args)
}

var helpTopics = map[string]func(*rootCmd, []string) error{
	"serve": func(r *rootCmd, args []string) error {
		_, err := parseServeCmd(r, args)
		return err
	},
	"user": func(r *rootCmd, args []string) error {
		cmd, err := parseUserCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"email": func(r *rootCmd, args []string) error {
		cmd, err := parseEmailCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"dlq": func(r *rootCmd, args []string) error {
		cmd, err := parseDlqCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"requests": func(r *rootCmd, args []string) error {
		cmd, err := parseRequestsCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"db": func(r *rootCmd, args []string) error {
		cmd, err := parseDbCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"perm": func(r *rootCmd, args []string) error {
		cmd, err := parsePermCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"board": func(r *rootCmd, args []string) error {
		cmd, err := parseBoardCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"blog": func(r *rootCmd, args []string) error {
		cmd, err := parseBlogCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"blogs": func(r *rootCmd, args []string) error {
		cmd, err := parseBlogCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"writing": func(r *rootCmd, args []string) error {
		cmd, err := parseWritingCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"news": func(r *rootCmd, args []string) error {
		cmd, err := parseNewsCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"announcement": func(r *rootCmd, args []string) error {
		cmd, err := parseAnnouncementCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"faq": func(r *rootCmd, args []string) error {
		cmd, err := parseFaqCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"ipban": func(r *rootCmd, args []string) error {
		cmd, err := parseIpBanCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"links": func(r *rootCmd, args []string) error {
		cmd, err := parseLinksCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"files": func(r *rootCmd, args []string) error {
		cmd, err := parseFilesCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"comment": func(r *rootCmd, args []string) error {
		cmd, err := parseCommentCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"comments": func(r *rootCmd, args []string) error {
		cmd, err := parseCommentCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"audit": func(r *rootCmd, args []string) error {
		cmd, err := parseAuditCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"notifications": func(r *rootCmd, args []string) error {
		cmd, err := parseNotificationsCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"repl": func(r *rootCmd, args []string) error {
		_, err := parseReplCmd(r, args)
		return err
	},
	"lang": func(r *rootCmd, args []string) error {
		cmd, err := parseLangCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"maintenance": func(r *rootCmd, args []string) error {
		cmd, err := parseMaintenanceCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"server": func(r *rootCmd, args []string) error {
		cmd, err := parseServerCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"config": func(r *rootCmd, args []string) error {
		cmd, err := parseConfigCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"page-size": func(r *rootCmd, args []string) error {
		cmd, err := parsePageSizeCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
	"subscription": func(r *rootCmd, args []string) error {
		cmd, err := parseSubscriptionCmd(r, args)
		if err != nil {
			return err
		}
		return cmd.Run()
	},
}

func (c *helpCmd) showHelp(args []string) error {
	if len(args) == 0 {
		c.rootCmd.fs.Usage()
		return nil
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	if runFn, ok := helpTopics[args[0]]; ok {
		err := runFn(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("%s: %w", args[0], err)
		}
		return nil
	}
	c.fs.Usage()
	return fmt.Errorf("unknown help topic %q", args[0])
}

func (c *helpCmd) Usage() {
	executeUsage(c.fs.Output(), "help_usage.txt", c)
}

func (c *helpCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*helpCmd)(nil)
