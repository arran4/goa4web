package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/runtimeconfig"
)

// serveCmd starts the web server.
type serveCmd struct {
	*rootCmd
	fs                *flag.FlagSet
	sessionSecret     string
	sessionSecretFile string
	args              []string
}

func parseServeCmd(parent *rootCmd, args []string) (*serveCmd, error) {
	c := &serveCmd{rootCmd: parent}
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.StringVar(&c.sessionSecret, "session-secret", "", "session secret key")
	fs.StringVar(&c.sessionSecretFile, "session-secret-file", "", "path to session secret file")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *serveCmd) Run() error {
	fileVals := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	app.ConfigFile = c.rootCmd.ConfigFile
	secretPath := c.sessionSecretFile
	if secretPath == "" {
		if v, ok := fileVals["SESSION_SECRET_FILE"]; ok {
			secretPath = v
		}
	}
	secret, err := core.LoadSessionSecret(core.OSFS{}, c.sessionSecret, secretPath)
	if err != nil {
		return fmt.Errorf("session secret: %w", err)
	}
	cfg := runtimeconfig.GenerateRuntimeConfig(c.rootCmd.fs, fileVals, os.Getenv)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.RunWithConfig(ctx, cfg, secret); err != nil {
		return err
	}
	return nil
}

// Usage prints command usage information with examples.
func (c *serveCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s serve [flags]\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
