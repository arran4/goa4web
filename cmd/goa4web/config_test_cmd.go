package main

import (
	"context"

	"flag"
	"fmt"
	"net/mail"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"

	coretemplates "github.com/arran4/goa4web/core/templates"
)

// configTestCmd implements "config test".
type configTestCmd struct {
	*configCmd
	fs *flag.FlagSet
}

func parseConfigTestCmd(parent *configCmd, args []string) (*configTestCmd, error) {
	c := &configTestCmd{configCmd: parent}
	c.fs = newFlagSet("test")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configTestCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing test command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "email":
		cmd, err := parseConfigTestEmailCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("email: %w", err)
		}
		return cmd.Run()
	case "db":
		cmd, err := parseConfigTestDBCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		return cmd.Run()
	case "dlq":
		cmd, err := parseConfigTestDLQCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("dlq: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown test command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *configTestCmd) Usage() {
	executeUsage(c.fs.Output(), "config_test_usage.txt", c)
}

func (c *configTestCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*configTestCmd)(nil)

type configTestEmailCmd struct {
	*configTestCmd
	fs *flag.FlagSet
}

func parseConfigTestEmailCmd(parent *configTestCmd, args []string) (*configTestEmailCmd, error) {
	c := &configTestEmailCmd{configTestCmd: parent}
	c.fs = newFlagSet("email")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configTestEmailCmd) Run() error {
	provider, err := c.rootCmd.emailReg.ProviderFromConfig(c.rootCmd.cfg)
	if err != nil || provider == nil {
		if err != nil {
			return fmt.Errorf("email provider error: %w", err)
		}
		return fmt.Errorf("email provider not configured")
	}
	var q db.Querier
	if conn, err := c.rootCmd.DB(); err == nil {
		q = db.New(conn)
	}
	ctx := context.Background()
	emails := config.GetAdminEmails(ctx, q, c.rootCmd.cfg)
	if len(emails) == 0 {
		return fmt.Errorf("no administrator emails configured")
	}
	htmlTmpls := coretemplates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := coretemplates.GetCompiledEmailTextTemplates(map[string]any{})
	for _, addrStr := range emails {
		toAddr := mail.Address{Address: addrStr}
		var buf strings.Builder
		if err := textTmpls.ExecuteTemplate(&buf, "testEmail.gotxt", nil); err != nil {
			return fmt.Errorf("exec text template: %w", err)
		}
		textBody := buf.String()
		buf.Reset()
		if err := htmlTmpls.ExecuteTemplate(&buf, "testEmail.gohtml", nil); err != nil {
			return fmt.Errorf("exec html template: %w", err)
		}
		var fromAddr mail.Address
		if f, err := mail.ParseAddress(c.rootCmd.cfg.EmailFrom); err == nil {
			fromAddr = *f
		} else {
			fromAddr = mail.Address{Address: c.rootCmd.cfg.EmailFrom}
		}
		msg, err := email.BuildMessage(fromAddr, toAddr, "Goa4Web Test Email", textBody, buf.String())
		if err != nil {
			return fmt.Errorf("build message: %w", err)
		}
		if err := provider.Send(ctx, toAddr, msg); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	return nil
}

type configTestDBCmd struct {
	*configTestCmd
	fs *flag.FlagSet
}

func parseConfigTestDBCmd(parent *configTestCmd, args []string) (*configTestDBCmd, error) {
	c := &configTestDBCmd{configTestCmd: parent}
	c.fs = newFlagSet("db")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configTestDBCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

type configTestDLQCmd struct {
	*configTestCmd
	fs *flag.FlagSet
}

func parseConfigTestDLQCmd(parent *configTestCmd, args []string) (*configTestDLQCmd, error) {
	c := &configTestDLQCmd{configTestCmd: parent}
	c.fs = newFlagSet("dlq")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configTestDLQCmd) Run() error {
	var q db.Querier
	if conn, err := c.rootCmd.DB(); err == nil {
		q = db.New(conn)
	}
	provider := c.rootCmd.dlqReg.ProviderFromConfig(c.rootCmd.cfg, q)
	if provider == nil {
		return fmt.Errorf("dlq provider not configured")
	}
	if err := provider.Record(context.Background(), "goa4web dlq test"); err != nil {
		return fmt.Errorf("dlq record: %w", err)
	}
	return nil
}
