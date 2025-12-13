package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/jmap"
)

type jmapTestSendCmd struct {
	*jmapCmd
	fs *flag.FlagSet
}

func parseJmapTestSendCmd(parent *jmapCmd, args []string) (*jmapTestSendCmd, error) {
	c := &jmapTestSendCmd{jmapCmd: parent}
	c.fs = newFlagSet("test-send")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jmapTestSendCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("missing recipient email address")
	}
	to, err := mail.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid recipient email address: %w", err)
	}

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

	provider := reg.ProviderFromConfig(cfg)
	if provider == nil {
		return fmt.Errorf("failed to create jmap provider")
	}

	fmt.Printf("Sending test email to %s\n", to.Address)

	from, err := mail.ParseAddress(cfg.EmailFrom)
	if err != nil {
		return fmt.Errorf("invalid from email address: %w", err)
	}

	raw, err := email.BuildMessage(*from, *to, "JMAP Test Email", "This is a test email from the goa4web jmap test-send command.", "")
	if err != nil {
		return fmt.Errorf("failed to create email message: %w", err)
	}

	if err := provider.Send(context.Background(), *to, raw); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Println("Email sent successfully")

	return nil
}

func (c *jmapTestSendCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_send_usage.txt", c)
}
