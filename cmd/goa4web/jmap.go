package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email/jmap"
)

// jmapCmd handles email-related subcommands.
type jmapCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseJmapCmd(parent *rootCmd, args []string) (*jmapCmd, error) {
	c := &jmapCmd{rootCmd: parent}
	c.fs = newFlagSet("jmap")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jmapCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing jmap command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "test":
		return c.runTest()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown jmap command %q", args[0])
	}
}

func (c *jmapCmd) runTest() error {
	cfg := c.cfg
	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		return fmt.Errorf("Email disabled: %s not set", config.EnvJMAPEndpoint)
	}
	acc := strings.TrimSpace(cfg.EmailJMAPAccount)
	id := strings.TrimSpace(cfg.EmailJMAPIdentity)

	httpClient := http.DefaultClient
	if cfg.EmailJMAPInsecure {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: tr}
	}

	if acc == "" || id == "" {
		var session *jmap.SessionResponse
		var err error
		for i := 0; i < 5; i++ {
			session, err = jmap.DiscoverSession(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass)
			if err == nil {
				break
			}
			fmt.Printf("Discovery attempt %d failed: %v. Retrying in 2s...\n", i+1, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			return fmt.Errorf("Email disabled: failed to discover JMAP session: %v", err)
		}
		b, _ := json.MarshalIndent(session, "", "  ")
		fmt.Printf("Discovered Session:\n%s\n", string(b))

		if acc == "" {
			acc = jmap.SelectAccountID(session)
			fmt.Printf("Selected AccountID: %s\n", acc)
		}
		if id == "" {
			id = jmap.SelectIdentityID(session)
			fmt.Printf("Selected IdentityID: %s\n", id)
		}
		if ep == "" {
			ep = session.APIURL
		}
		if id == "" && acc != "" {
			// Try to fetch identities via API
			fetchedId, err := jmap.DiscoverIdentityID(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
			if err != nil {
				fmt.Printf("Failed to discover Identity ID via API: %v\n", err)
			} else if fetchedId != "" {
				id = fetchedId
				fmt.Printf("Discovered Identity ID via API: %s\n", id)
			} else {
				fmt.Println("DiscoverIdentityID returned empty ID")
			}
		}
	}

	if acc == "" || id == "" {
		return fmt.Errorf("Email disabled: %s or %s not set and could not be discovered", config.EnvJMAPAccount, config.EnvJMAPIdentity)
	}

	provider := jmap.NewProvider(ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc, id, cfg.EmailFrom, httpClient)

	fmt.Printf("JMAP Provider Configured:\nEndpoint: %s\nUser: %s\nAccountID: %s\nIdentityID: %s\n", ep, cfg.EmailJMAPUser, acc, id)

	targetEmail := cfg.EmailJMAPUser // Send to self
	subject := fmt.Sprintf("JMAP Test Email %d", time.Now().Unix())
	body := "This is a test email sent from the JMAP CLI test command."

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", cfg.EmailFrom, targetEmail, subject, body)

	fmt.Printf("Sending email to %s with subject %q...\n", targetEmail, subject)
	err := provider.Send(context.Background(), mail.Address{Address: targetEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	fmt.Println("Email sent successfully.")

	fmt.Println("Waiting for email to arrive...")
	// Poll for email
	ctx := context.Background()
	inboxID, err := provider.GetInboxID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inbox ID: %w", err)
	}
	fmt.Printf("Inbox ID: %s. Checking inbox...\n", inboxID)

	for i := 0; i < 10; i++ {
		fmt.Printf("Attempt %d/10...\n", i+1)
		msgIDs, err := provider.QueryInbox(ctx, inboxID, 10)
		if err != nil {
			fmt.Printf("Error querying inbox: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(msgIDs) > 0 {
			emails, err := provider.GetMessages(ctx, msgIDs)
			if err != nil {
				fmt.Printf("Error getting messages: %v\n", err)
			} else {
				for _, email := range emails {
					if email.Subject == subject {
						fmt.Printf("SUCCESS: Found email '%s' (ID: %s) from %v received at %s\n", email.Subject, email.ID, email.From, email.ReceivedAt)
						return nil
					}
				}
				fmt.Println("Email not found in recent inbox messages yet.")
			}
		} else {
			fmt.Println("Inbox empty or query returned no results.")
		}

		// Fallback check: Check ANY message (debug)
		allIDs, err := provider.GetAllMessages(ctx, 5)
		if err == nil && len(allIDs) > 0 {
			allEmails, err := provider.GetMessages(ctx, allIDs)
			if err == nil {
				fmt.Println("Debug: Recent messages in account (ANY mailbox):")
				for _, e := range allEmails {
					fmt.Printf(" - ID: %s, Subj: %s, Recv: %s\n", e.ID, e.Subject, e.ReceivedAt)
					if e.Subject == subject {
						fmt.Println("   (This is the email we sent! It exists but not in inbox?)")
					}
				}
			}
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("timed out waiting for email")
}

// Usage prints command usage information with examples.
func (c *jmapCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_usage.txt", c)
}

func (c *jmapCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*jmapCmd)(nil)
