package main

import (
	"context"

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

// serveCmd starts the web server.
type serveCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseServeCmd(parent *rootCmd, args []string) (*serveCmd, error) {
	c := &serveCmd{rootCmd: parent}
	c.fs = config.NewRuntimeFlagSet("serve")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
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
       if err := app.RunWithConfig(ctx, cfg, secret, signKey, c.rootCmd.dbReg, c.rootCmd.emailReg, c.rootCmd.dlqReg); err != nil {
		return err
	}
	return nil
}

// Usage prints command usage information with examples.
func (c *serveCmd) Usage() {
	executeUsage(c.fs.Output(), "serve_usage.txt", c)
}

func (c *serveCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*serveCmd)(nil)
