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
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(c.fs),
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)

	c.rootCmd.Infof("Starting Goa4Web v%s (commit: %s; build date: %s)", version, commit, date)
	listenMsg := fmt.Sprintf("Listening on: %s", cfg.HTTPListen)
	if cfg.HTTPHostname != "" {
		listenMsg += fmt.Sprintf(" (Hostname: %s)", cfg.HTTPHostname)
	}
	c.rootCmd.Infof("%s", listenMsg)

	secret, err := config.LoadOrCreateSecret(core.OSFS{}, cfg.SessionSecret, cfg.SessionSecretFile, config.EnvSessionSecret, config.EnvSessionSecretFile)
	if err != nil {
		return fmt.Errorf("session secret: %w", err)
	}
	signKey, err := config.LoadOrCreateSecret(core.OSFS{}, cfg.ImageSignSecret, cfg.ImageSignSecretFile, config.EnvImageSignSecret, config.EnvImageSignSecretFile)
	if err != nil {
		return fmt.Errorf("image sign secret: %w", err)
	}
	linkKey, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cfg.LinkSignSecret, cfg.LinkSignSecretFile)
	if err != nil {
		return fmt.Errorf("link sign secret: %w", err)
	}
	apiKey, err := config.LoadOrCreateAdminAPISecret(core.OSFS{}, cfg.AdminAPISecret, cfg.AdminAPISecretFile)
	if err != nil {
		return fmt.Errorf("admin api secret: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv, err := app.NewServer(ctx, cfg, c.rootCmd.adminHandlers,
		app.WithSessionSecret(secret),
		app.WithImageSignSecret(signKey),
		app.WithLinkSignSecret(linkKey),
		app.WithDBRegistry(c.rootCmd.dbReg),
		app.WithEmailRegistry(c.rootCmd.emailReg),
		app.WithDLQRegistry(c.rootCmd.dlqReg),
		app.WithTasksRegistry(c.rootCmd.tasksReg),
		app.WithAPISecret(apiKey),
		app.WithRouterRegistry(c.rootCmd.routerReg),
	)
	if err != nil {
		return err
	}
	defer srv.Close()
	if err := srv.RunContext(ctx); err != nil {
		return err
	}
	return nil
}

// Usage prints command usage information with examples.
func (c *serveCmd) Usage() {
	executeUsage(c.fs.Output(), "serve_usage.txt", c)
}

func (c *serveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*serveCmd)(nil)
