package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"time"

	"github.com/arran4/goa4web/internal/email/jmap"
)

// jmapTestSendCmd handles the 'jmap test-send' subcommand.
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
	info, err := c.discoverJmapSession()
	if err != nil {
		return err
	}

	cfg := c.cfg
	provider := jmap.NewProvider(info.APIEndpoint, cfg.EmailJMAPUser, cfg.EmailJMAPPass, info.AccountID, info.IdentityID, cfg.EmailFrom, info.Client)

	c.rootCmd.Infof("JMAP Provider Configured:\n  Endpoint: %s\n  User: %s\n  AccountID: %s\n  IdentityID: %s\n", info.APIEndpoint, cfg.EmailJMAPUser, info.AccountID, info.IdentityID)

	targetEmail := cfg.EmailJMAPUser // Send to self
	subject := fmt.Sprintf("JMAP Test Email %d", time.Now().Unix())
	body := "This is a test email sent from the JMAP CLI test command."

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", cfg.EmailFrom, targetEmail, subject, body)

	c.rootCmd.Infof("Sending email to %s with subject %q...\n", targetEmail, subject)
	err = provider.Send(context.Background(), mail.Address{Address: targetEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	c.rootCmd.Infof("Email sent successfully.\n")

	c.rootCmd.Infof("Waiting for email to arrive...\n")
	ctx := context.Background()
	inboxID, err := provider.GetInboxID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inbox ID: %w", err)
	}
	c.rootCmd.Infof("Inbox ID: %s. Checking inbox...\n", inboxID)

	for i := 0; i < 10; i++ {
		c.rootCmd.Infof("Attempt %d/10...\n", i+1)
		msgIDs, err := provider.QueryInbox(ctx, inboxID, 10)
		if err != nil {
			c.rootCmd.Infof("Error querying inbox: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(msgIDs) > 0 {
			emails, err := provider.GetMessages(ctx, msgIDs)
			if err != nil {
				c.rootCmd.Infof("Error getting messages: %v\n", err)
			} else {
				for _, email := range emails {
					if email.Subject == subject {
						c.rootCmd.Infof("SUCCESS: Found email '%s' (ID: %s) from %v received at %s\n", email.Subject, email.ID, email.From, email.ReceivedAt)
						return nil
					}
				}
				c.rootCmd.Infof("Email not found in recent inbox messages yet.\n")
			}
		} else {
			c.rootCmd.Infof("Inbox empty or query returned no results.\n")
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("timed out waiting for email")
}

// Usage prints command usage information.
func (c *jmapTestSendCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_send_usage.txt", c)
}

func (c *jmapTestSendCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*jmapTestSendCmd)(nil)
