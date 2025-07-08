package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

// emailQueueResendCmd implements "email queue resend".
type emailQueueResendCmd struct {
	*emailQueueCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseEmailQueueResendCmd(parent *emailQueueCmd, args []string) (*emailQueueResendCmd, error) {
	c := &emailQueueResendCmd{emailQueueCmd: parent}
	fs := flag.NewFlagSet("resend", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "pending email id")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailQueueResendCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	e, err := queries.GetPendingEmailByID(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get email: %w", err)
	}
	provider := email.ProviderFromConfig(c.rootCmd.cfg)
	if provider != nil {
		msg, err := email.BuildMessage(c.rootCmd.cfg.EmailFrom, e.ToEmail, e.Subject, e.Body, "")
		if err != nil {
			return fmt.Errorf("build message: %w", err)
		}
		if err := provider.Send(ctx, e.ToEmail, e.Subject, msg); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	if err := queries.MarkEmailSent(ctx, e.ID); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}
	return nil
}
