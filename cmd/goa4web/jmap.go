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

		// Determine the API Endpoint to use
		apiURL := session.APIURL
		if override := strings.TrimSpace(cfg.EmailJMAPEndpointOverride); override != "" {
			fmt.Printf("Using configured JMAP Endpoint Override: %s\n", override)
			apiURL = override
			ep = override
		} else if ep == "" {
			ep = session.APIURL
		} else {
			// If ep was set (discovery endpoint), check if it has a path that suggests it is the API endpoint
			// Same logic as in providerFromConfig
			// But for test command, let's trust session.APIURL unless overridden or if ep looks like a full URL
			if session.APIURL == "" {
				// Should not happen if valid session
				apiURL = ep
			}
		}

		if id == "" && acc != "" {
			// Try to fetch identities via API
			fmt.Printf("Attempting to discover identity via API: %s\n", apiURL)
			fetchedId, err := jmap.DiscoverIdentityID(context.Background(), httpClient, apiURL, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
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

	// Double check endpoint for provider init
	if override := strings.TrimSpace(cfg.EmailJMAPEndpointOverride); override != "" {
		ep = override
	}

	provider := jmap.NewProvider(ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc, id, cfg.EmailFrom, httpClient)
	jmapProvider, ok := provider.(*jmap.Provider)
	if !ok {
		return fmt.Errorf("internal error: failed to assert JMAP provider type")
	}
	fmt.Printf("JMAP Provider Configured:\nEndpoint: %s\nUser: %s\nAccountID: %s\nIdentityID: %s\n", ep, cfg.EmailJMAPUser, acc, id)

	targetEmail := cfg.EmailJMAPUser // Send to self
	subject := fmt.Sprintf("JMAP Test Email %d", time.Now().Unix())
	body := "This is a test email sent from the JMAP CLI test command."

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", cfg.EmailFrom, targetEmail, subject, body)

	fmt.Printf("Sending email to %s with subject %q...\n", targetEmail, subject)
	err := jmapProvider.Send(context.Background(), mail.Address{Address: targetEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	fmt.Println("Email sent successfully.")

	fmt.Println("Waiting for email to arrive...")
	// Poll for email
	ctx := context.Background()
	inboxID, err := jmapProvider.GetInboxID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inbox ID: %w", err)
	}
	fmt.Printf("Inbox ID: %s. Checking inbox...\n", inboxID)

	for i := 0; i < 10; i++ {
		fmt.Printf("Attempt %d/10...\n", i+1)
		msgIDs, err := jmapProvider.QueryInbox(ctx, inboxID, 10)
		if err != nil {
			fmt.Printf("Error querying inbox: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(msgIDs) > 0 {
			emails, err := jmapProvider.GetMessages(ctx, msgIDs)
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
		allIDs, err := jmapProvider.GetAllMessages(ctx, 5)
		if err == nil && len(allIDs) > 0 {
			allEmails, err := jmapProvider.GetMessages(ctx, allIDs)
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
