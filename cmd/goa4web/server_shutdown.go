package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/adminapi"
)

// serverShutdownCmd implements "server shutdown".
type serverShutdownCmd struct {
	*serverCmd
	fs      *flag.FlagSet
	Timeout time.Duration
	Mode    string
}

func parseServerShutdownCmd(parent *serverCmd, args []string) (*serverShutdownCmd, error) {
	c := &serverShutdownCmd{serverCmd: parent}
	c.fs = newFlagSet("shutdown")
	c.fs.DurationVar(&c.Timeout, "timeout", 5*time.Second, "shutdown timeout")
	c.fs.StringVar(&c.Mode, "mode", "", "shutdown mode (rest or local)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *serverShutdownCmd) Run() error {
	mode := c.Mode
	if mode == "" {
		if c.rootCmd.adminHandlers.Srv == nil {
			mode = "rest"
		} else {
			mode = "local"
		}
	}
	switch mode {
	case "local":
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()
		if err := c.rootCmd.adminHandlers.Srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	case "rest":
		return c.restShutdown()
	default:
		return fmt.Errorf("invalid mode %q", mode)
	}
}

func (c *serverShutdownCmd) restShutdown() error {
	cfg := c.rootCmd.cfg
	key, err := config.LoadOrCreateAdminAPISecret(core.OSFS{}, cfg.AdminAPISecret, cfg.AdminAPISecretFile)
	if err != nil {
		return fmt.Errorf("admin api secret: %w", err)
	}
	signer := adminapi.NewSigner(key)
	ts, sig := signer.Sign(http.MethodPost, "/admin/api/shutdown")
	req, err := http.NewRequest(http.MethodPost, cfg.BaseURL+"/admin/api/shutdown", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("shutdown request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shutdown status %s", resp.Status)
	}
	return nil
}
