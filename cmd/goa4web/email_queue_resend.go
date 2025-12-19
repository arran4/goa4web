package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// emailQueueResendCmd implements "email queue resend".
type emailQueueResendCmd struct {
	*emailQueueCmd
	fs *flag.FlagSet
	ID int
}

func parseEmailQueueResendCmd(parent *emailQueueCmd, args []string) (*emailQueueResendCmd, error) {
	c := &emailQueueResendCmd{emailQueueCmd: parent}
	c.fs = newFlagSet("resend")
	c.fs.IntVar(&c.ID, "id", 0, "pending email id")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailQueueResendCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	e, err := queries.AdminGetPendingEmailByID(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get email: %w", err)
	}
	provider, err := c.rootCmd.emailReg.ProviderFromConfig(c.rootCmd.cfg)
	if err != nil {
		return fmt.Errorf("email provider: %w", err)
	}
	if provider != nil {
		addr, err := emailqueue.ResolveQueuedEmailAddress(ctx, queries, c.rootCmd.cfg, &db.SystemListPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail})
		if err != nil {
			return err
		}
		if err := provider.Send(ctx, addr, []byte(e.Body)); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	if err := queries.SystemMarkPendingEmailSent(ctx, e.ID); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}
	return nil
}
