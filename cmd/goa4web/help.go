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

func (c *helpCmd) showHelp(args []string) error {
	if len(args) == 0 {
		c.rootCmd.fs.Usage()
		return nil
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "serve":
		_, err := parseServeCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	case "user":
		cmd, err := parseUserCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("user: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "email":
		cmd, err := parseEmailCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("email: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "requests":
		cmd, err := parseRequestsCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("requests: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "db":
		cmd, err := parseDbCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("db: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "perm":
		cmd, err := parsePermCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("perm: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "board":
		cmd, err := parseBoardCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("board: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "blog", "blogs":
		cmd, err := parseBlogCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("blog: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "writing":
		cmd, err := parseWritingCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("writing: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "news":
		cmd, err := parseNewsCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("news: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "announcement":
		cmd, err := parseAnnouncementCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("announcement: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "faq":
		cmd, err := parseFaqCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("faq: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "ipban":
		cmd, err := parseIpBanCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("ipban: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "links":
		cmd, err := parseLinksCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("links: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "files":
		cmd, err := parseFilesCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("files: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "comment", "comments":
		cmd, err := parseCommentCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("comment: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "audit":
		cmd, err := parseAuditCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("audit: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "notifications":
		cmd, err := parseNotificationsCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("notifications: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "repl":
		_, err := parseReplCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("repl: %w", err)
		}
		return nil
	case "lang":
		cmd, err := parseLangCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("lang: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "maintenance":
		cmd, err := parseMaintenanceCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("maintenance: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "server":
		cmd, err := parseServerCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("server: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "config":
		cmd, err := parseConfigCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("config: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "page-size":
		cmd, err := parsePageSizeCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("page-size: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	case "subscription":
		cmd, err := parseSubscriptionCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("subscription: %w", err)
		}
		if err == nil {
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown help topic %q", args[0])
	}
}

func (c *helpCmd) Usage() {
	executeUsage(c.fs.Output(), "help_usage.txt", c)
}

func (c *helpCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*helpCmd)(nil)
