package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/jmap"
	"github.com/arran4/goa4web/internal/email/log"
	"github.com/arran4/goa4web/internal/email/sendgrid"
	"github.com/arran4/goa4web/internal/email/ses"
	"github.com/arran4/goa4web/internal/email/smtp"
)

// emailTestCmd handles the `email test` subcommand.
type emailTestCmd struct {
	*emailCmd
	fs *flag.FlagSet
}

func parseEmailTestCmd(parent *emailCmd, args []string) (*emailTestCmd, error) {
	c := &emailTestCmd{emailCmd: parent}
	c.fs = newFlagSet("test")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailTestCmd) Run() error {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config file: %w", err)
	}
	cfg := config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)

	reg := email.NewRegistry()
	jmap.Register(reg)
	log.Register(reg)
	ses.Register(reg)
	sendgrid.Register(reg)
	smtp.Register(reg)

	provider, err := reg.ProviderFromConfig(cfg)
	if err != nil || provider == nil {
		if err != nil {
			return fmt.Errorf("failed to create email provider for %q: %w", cfg.EmailProvider, err)
		}
		return fmt.Errorf("failed to create email provider for %q", cfg.EmailProvider)
	}

	c.Infof("Testing email provider %q", cfg.EmailProvider)

	if err := provider.TestConfig(context.Background()); err != nil {
		return fmt.Errorf("email provider test failed: %w", err)
	}

	c.Infof("Email provider test successful")

	return nil
}

// Usage prints the command's usage information.
func (c *emailTestCmd) Usage() {
	executeUsage(c.fs.Output(), "email_test_usage.txt", c)
}

func (c *emailTestCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailTestCmd)(nil)
