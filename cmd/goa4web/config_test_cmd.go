package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/mail"
	"strings"
	"text/template"

	coretemplates "github.com/arran4/goa4web/core/templates"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/utils/emailutil"
)

//go:embed templates/config_test_usage.txt
var configTestUsageTemplate string

// configTestCmd implements "config test".
type configTestCmd struct {
	*configCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigTestCmd(parent *configCmd, args []string) (*configTestCmd, error) {
	c := &configTestCmd{configCmd: parent}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configTestCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing test command")
	}
	switch c.args[0] {
	case "email":
		cmd, err := parseConfigTestEmailCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("email: %w", err)
		}
		return cmd.Run()
	case "db":
		cmd, err := parseConfigTestDBCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		return cmd.Run()
	case "dlq":
		cmd, err := parseConfigTestDLQCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("dlq: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown test command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *configTestCmd) Usage() {
	executeUsage(c.fs.Output(), configTestUsageTemplate, c.fs, c.rootCmd.fs.Name())
}

type configTestEmailCmd struct {
	*configTestCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigTestEmailCmd(parent *configTestCmd, args []string) (*configTestEmailCmd, error) {
	c := &configTestEmailCmd{configTestCmd: parent}
	fs := flag.NewFlagSet("email", flag.ContinueOnError)
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configTestEmailCmd) Run() error {
	provider := email.ProviderFromConfig(c.rootCmd.cfg)
	if provider == nil {
		return fmt.Errorf("email provider not configured")
	}
	var q *dbpkg.Queries
	if db, err := c.rootCmd.DB(); err == nil {
		q = dbpkg.New(db)
	}
	ctx := context.Background()
	emails := emailutil.GetAdminEmails(ctx, q)
	if len(emails) == 0 {
		return fmt.Errorf("no administrator emails configured")
	}
	for _, addrStr := range emails {
		toAddr := mail.Address{Address: addrStr}
		var buf strings.Builder
		t, err := template.New("txt").Parse(coretemplates.TestEmailText)
		if err != nil {
			return fmt.Errorf("parse text template: %w", err)
		}
		if err := t.Execute(&buf, nil); err != nil {
			return fmt.Errorf("exec text template: %w", err)
		}
		textBody := buf.String()
		buf.Reset()
		ht, err := template.New("html").Parse(coretemplates.TestEmailHTML)
		if err != nil {
			return fmt.Errorf("parse html template: %w", err)
		}
		if err := ht.Execute(&buf, nil); err != nil {
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
	fs   *flag.FlagSet
	args []string
}

func parseConfigTestDBCmd(parent *configTestCmd, args []string) (*configTestDBCmd, error) {
	c := &configTestDBCmd{configTestCmd: parent}
	fs := flag.NewFlagSet("db", flag.ContinueOnError)
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configTestDBCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

type configTestDLQCmd struct {
	*configTestCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigTestDLQCmd(parent *configTestCmd, args []string) (*configTestDLQCmd, error) {
	c := &configTestDLQCmd{configTestCmd: parent}
	fs := flag.NewFlagSet("dlq", flag.ContinueOnError)
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configTestDLQCmd) Run() error {
	var q *dbpkg.Queries
	if db, err := c.rootCmd.DB(); err == nil {
		q = dbpkg.New(db)
	}
	provider := dlq.ProviderFromConfig(c.rootCmd.cfg, q)
	if provider == nil {
		return fmt.Errorf("dlq provider not configured")
	}
	if err := provider.Record(context.Background(), "goa4web dlq test"); err != nil {
		return fmt.Errorf("dlq record: %w", err)
	}
	return nil
}
