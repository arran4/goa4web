package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email/jmap"
)

// jmapSessionInfo holds discovered JMAP session details.
type jmapSessionInfo struct {
	Session     *jmap.SessionResponse
	Client      *http.Client
	AccountID   string
	IdentityID  string
	APIEndpoint string
}

// discoverJmapSession attempts to discover JMAP session information.
func (c *jmapCmd) discoverJmapSession() (*jmapSessionInfo, error) {
	cfg := c.cfg
	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		return nil, fmt.Errorf("email disabled: %s not set", config.EnvJMAPEndpoint)
	}

	httpClient := http.DefaultClient
	if cfg.EmailJMAPInsecure {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: tr}
	}

	acc := strings.TrimSpace(cfg.EmailJMAPAccount)
	id := strings.TrimSpace(cfg.EmailJMAPIdentity)

	var session *jmap.SessionResponse
	var err error
	if acc == "" || id == "" {
		for i := 0; i < 5; i++ {
			session, err = jmap.DiscoverSession(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass)
			if err == nil {
				break
			}
			c.rootCmd.Infof("Discovery attempt %d failed: %v. Retrying in 2s...\n", i+1, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			return nil, fmt.Errorf("email disabled: failed to discover JMAP session: %w", err)
		}

		if acc == "" {
			acc = jmap.SelectAccountID(session)
		}
		if id == "" {
			id = jmap.SelectIdentityID(session)
		}
	}

	// Determine the API Endpoint to use
	apiURL := ""
	if session != nil {
		apiURL = session.APIURL
	}
	if override := strings.TrimSpace(cfg.EmailJMAPEndpointOverride); override != "" {
		apiURL = override
		ep = override
	} else if ep == "" && session != nil {
		ep = session.APIURL
	}

	if id == "" && acc != "" && apiURL != "" {
		c.rootCmd.Infof("Attempting to discover identity via API: %s\n", apiURL)
		fetchedId, err := jmap.DiscoverIdentityID(context.Background(), httpClient, apiURL, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
		if err != nil {
			c.rootCmd.Infof("Failed to discover Identity ID via API: %v\n", err)
		} else if fetchedId != "" {
			id = fetchedId
		}
	}

	if acc == "" || id == "" {
		return nil, fmt.Errorf("email disabled: %s or %s not set and could not be discovered", config.EnvJMAPAccount, config.EnvJMAPIdentity)
	}

	return &jmapSessionInfo{
		Session:     session,
		Client:      httpClient,
		AccountID:   acc,
		IdentityID:  id,
		APIEndpoint: ep,
	}, nil
}

// runTestConfig implements the 'test-config' subcommand.
func (c *jmapCmd) runTestConfig() error {
	info, err := c.discoverJmapSession()
	if err != nil {
		return err
	}

	if info.Session != nil {
		b, _ := json.MarshalIndent(info.Session, "", "  ")
		fmt.Printf("Discovered Session:\n%s\n", string(b))
	}
	fmt.Printf("Selected AccountID: %s\n", info.AccountID)
	fmt.Printf("Selected IdentityID: %s\n", info.IdentityID)
	fmt.Printf("Using API Endpoint: %s\n", info.APIEndpoint)
	fmt.Println("\nJMAP configuration appears to be valid.")
	return nil
}

// runTestSend implements the 'test-send' subcommand.
func (c *jmapCmd) runTestSend() error {
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
