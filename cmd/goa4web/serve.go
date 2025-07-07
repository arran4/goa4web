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
	sopts := []runtimeconfig.StringOption{
		{Name: "session-secret", Env: config.EnvSessionSecret, Usage: "session secret key"},
		{Name: "session-secret-file", Env: config.EnvSessionSecretFile, Usage: "path to session secret file"},
	}
	fs := runtimeconfig.NewRuntimeFlagSetWithOptions("serve", sopts, nil)
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.sessionSecret = fs.Lookup("session-secret").Value.String()
	c.sessionSecretFile = fs.Lookup("session-secret-file").Value.String()
	c.args = fs.Args()
	return c, nil
}

func (c *serveCmd) Run() error {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config file: %w", err)
	}
	app.ConfigFile = c.rootCmd.ConfigFile
	secretPath := c.sessionSecretFile
	if secretPath == "" {
		if v, ok := fileVals["SESSION_SECRET_FILE"]; ok {
			secretPath = v
		}
	}
	secret, err := core.LoadSessionSecret(core.OSFS{}, c.sessionSecret, secretPath, config.EnvSessionSecret, config.EnvSessionSecretFile)
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
