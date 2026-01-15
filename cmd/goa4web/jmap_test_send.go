package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"time"

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
	ep, acc, id, httpClient, err := discoverJmapLogic(c.jmapCmd)
	if err != nil {
		return err
	}
	cfg := c.cfg
	provider := jmap.NewProvider(ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc, id, cfg.EmailFrom, httpClient)

	targetEmail := cfg.EmailJMAPUser // Send to self
	subject := fmt.Sprintf("JMAP Test Email %d", time.Now().Unix())
	body := "This is a test email sent from the JMAP CLI test command."

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", cfg.EmailFrom, targetEmail, subject, body)

	c.Infof("Sending email to %s with subject %q...\n", targetEmail, subject)
	err = provider.Send(context.Background(), mail.Address{Address: targetEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	c.Infof("Email sent successfully.\n")

	c.Infof("Waiting for email to arrive...\n")
	// Poll for email
	ctx := context.Background()
	inboxID, err := provider.GetInboxID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inbox ID: %w", err)
	}
	c.Infof("Inbox ID: %s. Checking inbox...\n", inboxID)

	for i := 0; i < 10; i++ {
		c.Infof("Attempt %d/10...\n", i+1)
		msgIDs, err := provider.QueryInbox(ctx, inboxID, 10)
		if err != nil {
			c.Verbosef("Error querying inbox: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(msgIDs) > 0 {
			emails, err := provider.GetMessages(ctx, msgIDs)
			if err != nil {
				c.Verbosef("Error getting messages: %v\n", err)
			} else {
				for _, email := range emails {
					if email.Subject == subject {
						c.Infof("SUCCESS: Found email '%s' (ID: %s) from %v received at %s\n", email.Subject, email.ID, email.From, email.ReceivedAt)
						return nil
					}
				}
				c.Verbosef("Email not found in recent inbox messages yet.\n")
			}
		} else {
			c.Verbosef("Inbox empty or query returned no results.\n")
		}
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("timed out waiting for email")
}

func (c *jmapTestSendCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_send_usage.txt", c)
}
