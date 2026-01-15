package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email/jmap"
)

func discoverJmapLogic(c *jmapCmd) (string, string, string, *http.Client, error) {
	cfg := c.cfg
	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		return "", "", "", nil, fmt.Errorf("Email disabled: %s not set", config.EnvJMAPEndpoint)
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
			c.Infof("Discovery attempt %d failed: %v. Retrying in 2s...\n", i+1, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			return "", "", "", nil, fmt.Errorf("Email disabled: failed to discover JMAP session: %v", err)
		}
		b, _ := json.MarshalIndent(session, "", "  ")
		c.Infof("Discovered Session:\n%s\n", string(b))

		if acc == "" {
			acc = jmap.SelectAccountID(session)
			c.Infof("Selected AccountID: %s\n", acc)
		}
		if id == "" {
			id = jmap.SelectIdentityID(session)
			c.Infof("Selected IdentityID: %s\n", id)
		}
		if ep == "" {
			ep = session.APIURL
		}
		if id == "" && acc != "" {
			// Try to fetch identities via API
			fetchedId, err := jmap.DiscoverIdentityID(context.Background(), httpClient, ep, cfg.EmailJMAPUser, cfg.EmailJMAPPass, acc)
			if err != nil {
				c.Verbosef("Failed to discover Identity ID via API: %v\n", err)
			} else if fetchedId != "" {
				id = fetchedId
				c.Infof("Discovered Identity ID via API: %s\n", id)
			} else {
				c.Verbosef("DiscoverIdentityID returned empty ID")
			}
		}
	}

	if acc == "" || id == "" {
		return "", "", "", nil, fmt.Errorf("Email disabled: %s or %s not set and could not be discovered", config.EnvJMAPAccount, config.EnvJMAPIdentity)
	}
	return ep, acc, id, httpClient, nil
}
