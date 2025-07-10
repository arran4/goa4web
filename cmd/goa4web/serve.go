package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/app"
)

//go:embed templates/serve_usage.txt
var serveUsageTemplate string

// serveCmd starts the web server.
type serveCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseServeCmd(parent *rootCmd, args []string) (*serveCmd, error) {
	c := &serveCmd{rootCmd: parent}
	fs := config.NewRuntimeFlagSet("serve")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *serveCmd) Run() error {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		if errors.Is(err, config.ErrConfigFileNotFound) {
			return fmt.Errorf("config file not found: %s", c.rootCmd.ConfigFile)
		}
		return fmt.Errorf("load config file: %w", err)
	}
	app.ConfigFile = c.rootCmd.ConfigFile
	cfg := config.GenerateRuntimeConfig(c.fs, fileVals, os.Getenv)
	secret, err := config.LoadOrCreateSecret(core.OSFS{}, cfg.SessionSecret, cfg.SessionSecretFile, config.EnvSessionSecret, config.EnvSessionSecretFile)
	if err != nil {
		return fmt.Errorf("session secret: %w", err)
	}
	signKey, err := config.LoadOrCreateSecret(core.OSFS{}, cfg.ImageSignSecret, cfg.ImageSignSecretFile, config.EnvImageSignSecret, config.EnvImageSignSecretFile)
	if err != nil {
		return fmt.Errorf("image sign secret: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.RunWithConfig(ctx, cfg, secret, signKey); err != nil {
		return err
	}
	return nil
}

// Usage prints command usage information with examples.
func (c *serveCmd) Usage() {
	executeUsage(c.fs.Output(), serveUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
