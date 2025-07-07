package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s config test <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  email\tsend a test email to administrators")
	fmt.Fprintln(w, "  db\ttest database connectivity")
	fmt.Fprintln(w, "  dlq\ttest dead letter queue")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s config test email\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config test db\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config test dlq\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
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
	for _, addr := range emails {
		if err := provider.Send(ctx, addr, "goa4web test", "goa4web email configuration works", ""); err != nil {
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
