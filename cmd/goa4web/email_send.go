package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/jmap"
	"github.com/arran4/goa4web/internal/email/log"
	"github.com/arran4/goa4web/internal/email/sendgrid"
	"github.com/arran4/goa4web/internal/email/ses"
	"github.com/arran4/goa4web/internal/email/smtp"
)

// emailSendCmd handles the `email send` subcommand.
type emailSendCmd struct {
	*emailCmd
	fs      *flag.FlagSet
	to      string
	subject string
	body    string
}

func parseEmailSendCmd(parent *emailCmd, args []string) (*emailSendCmd, error) {
	c := &emailSendCmd{emailCmd: parent}
	c.fs = newFlagSet("send")
	c.fs.StringVar(&c.to, "to", "", "The recipient's email address.")
	c.fs.StringVar(&c.subject, "subject", "", "The subject of the email.")
	c.fs.StringVar(&c.body, "body", "", "The body of the email.")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailSendCmd) Run() error {
	if c.to == "" {
		return fmt.Errorf("missing -to recipient email address")
	}
	to, err := mail.ParseAddress(c.to)
	if err != nil {
		return fmt.Errorf("invalid recipient email address: %w", err)
	}

	cfg, err := c.rootCmd.RuntimeConfig()
	if err != nil {
		return err
	}

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

	c.Infof("Sending test email to %s", to.Address)

	from, err := mail.ParseAddress(cfg.EmailFrom)
	if err != nil {
		return fmt.Errorf("invalid from email address %q: %w", cfg.EmailFrom, err)
	}

	raw, err := email.BuildMessage(*from, *to, c.subject, c.body, "")
	if err != nil {
		return fmt.Errorf("failed to create email message: %w", err)
	}

	if err := provider.Send(context.Background(), *to, raw); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.Infof("Email sent successfully")

	return nil
}

// Usage prints the command's usage information.
func (c *emailSendCmd) Usage() {
	executeUsage(c.fs.Output(), "email_send_usage.txt", c)
}
