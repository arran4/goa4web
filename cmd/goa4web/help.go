package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/help_usage.txt
var helpUsageTemplate string

// helpCmd displays usage information for commands.
type helpCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseHelpCmd(parent *rootCmd, args []string) (*helpCmd, error) {
	c := &helpCmd{rootCmd: parent}
	fs := flag.NewFlagSet("help", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *helpCmd) Run() error {
	if len(c.args) == 0 {
		c.rootCmd.fs.Usage()
		return nil
	}
	return c.showHelp(c.args)
}

func (c *helpCmd) showHelp(args []string) error {
	if len(args) == 0 {
		c.rootCmd.fs.Usage()
		return nil
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
			_ = cmd.Run()
		}
		return nil
	case "email":
		cmd, err := parseEmailCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("email: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "db":
		cmd, err := parseDbCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("db: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "perm":
		cmd, err := parsePermCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("perm: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "board":
		cmd, err := parseBoardCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("board: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "blog", "blogs":
		cmd, err := parseBlogCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("blog: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "writing", "writings":
		cmd, err := parseWritingCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("writing: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "news":
		cmd, err := parseNewsCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("news: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "faq":
		cmd, err := parseFaqCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("faq: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "ipban":
		cmd, err := parseIpBanCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("ipban: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "audit":
		cmd, err := parseAuditCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("audit: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "lang":
		cmd, err := parseLangCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("lang: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "server":
		cmd, err := parseServerCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("server: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "password":
		cmd, err := parsePasswordCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("password: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	case "config":
		cmd, err := parseConfigCmd(c.rootCmd, append(args[1:], "-h"))
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("config: %w", err)
		}
		if err == nil {
			_ = cmd.Run()
		}
		return nil
	default:
		return fmt.Errorf("unknown help topic %q", args[0])
	}
}

func (c *helpCmd) Usage() {
	executeUsage(c.fs.Output(), helpUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
