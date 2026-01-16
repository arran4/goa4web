package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
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

// jmapSessionInfo holds discovered JMAP session details.
type jmapSessionInfo struct {
	Session     *jmap.SessionResponse
	Client      *http.Client
	AccountID   string
	IdentityID  string
	APIEndpoint string
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
	case "test-config":
		cmd, err := parseJmapTestConfigCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "test-send":
		cmd, err := parseJmapTestSendCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown jmap command %q", args[0])
	}
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
  
  if len(msgIDs) > 0 {
    fmt.Printf("Found %d recent messages, checking subjects...\n", len(msgIDs))
    emails, err := provider.GetMessages(ctx, msgIDs)
    if err != nil {
      fmt.Printf("Error getting messages: %v\n", err)
    } else {
      for _, email := range emails {
        fmt.Printf(" - Checking email with subject: %q\n", email.Subject)
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

	return &jmapSessionInfo{
		Session:     session,
		Client:      httpClient,
		AccountID:   acc,
		IdentityID:  id,
		APIEndpoint: ep,
	}, nil
}

func (c *jmapCmd) printSessionInfo(info *jmapSessionInfo) error {
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

// Usage prints command usage information with examples.
func (c *jmapCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_usage.txt", c)
}

func (c *jmapCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*jmapCmd)(nil)
