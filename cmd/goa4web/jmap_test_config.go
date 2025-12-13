package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/email/jmap"
)

type jmapTestConfigCmd struct {
	*jmapCmd
	fs *flag.FlagSet
}

func parseJmapTestConfigCmd(parent *jmapCmd, args []string) (*jmapTestConfigCmd, error) {
	c := &jmapTestConfigCmd{jmapCmd: parent}
	c.fs = newFlagSet("test-config")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jmapTestConfigCmd) Run() error {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config file: %w", err)
	}
	cfg := config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)

	ep := strings.TrimSpace(cfg.EmailJMAPEndpoint)
	if ep == "" {
		return fmt.Errorf("jmap endpoint not set")
	}

	httpClient := http.DefaultClient
	if cfg.EmailJMAPInsecure {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = &http.Client{Transport: tr}
	}

	fmt.Printf("Performing JMAP discovery for endpoint: %s\n", ep)

	wellKnown, err := jmap.JmapWellKnownURL(ep)
	if err != nil {
		return fmt.Errorf("failed to construct well-known URL: %w", err)
	}

	fmt.Printf("Well-known URL: %s\n", wellKnown)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, wellKnown, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if cfg.EmailJMAPUser != "" || cfg.EmailJMAPPass != "" {
		fmt.Println("Using basic authentication")
		req.SetBasicAuth(cfg.EmailJMAPUser, cfg.EmailJMAPPass)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Println("Response body:")
	fmt.Println(string(body))

	if resp.StatusCode >= 300 {
		return fmt.Errorf("jmap session discovery failed: %s", resp.Status)
	}

	var session jmap.SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return fmt.Errorf("failed to decode session response: %w", err)
	}

	acc := jmap.SelectAccountID(&session)
	id := jmap.SelectIdentityID(&session)

	fmt.Printf("Discovered Account ID: %s\n", acc)
	fmt.Printf("Discovered Identity ID: %s\n", id)

	return nil
}

func (c *jmapTestConfigCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_config_usage.txt", c)
}
