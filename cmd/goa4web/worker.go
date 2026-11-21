package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/workers"
)

// workerShutdownTimeout bounds cooperative durable worker shutdown after SIGINT/SIGTERM.
const workerShutdownTimeout = 5 * time.Second

// workerCmd runs standalone durable background job workers sharing the configured database.
type workerCmd struct {
	*rootCmd
	fs       *flag.FlagSet
	workerID string
}

func parseWorkerCmd(parent *rootCmd, args []string) (*workerCmd, error) {
	c := &workerCmd{rootCmd: parent}
	c.fs = config.NewRuntimeFlagSet("worker")
	c.fs.StringVar(&c.workerID, "worker-id", "", "Optional worker lease identifier prefix")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *workerCmd) Run() error {
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(c.fs),
		config.WithFileValues(c.ConfigFileValues),
		config.WithGetenv(os.Getenv),
	)
	c.cfg = cfg

	sdb, err := c.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.NewForDriver(sdb, cfg.DBDriver)
	svc := jobs.NewConfiguredService(q)

	ctx, stop := notifyShutdownContext(context.Background())
	defer stop()

	c.Infof("Starting standalone durable worker (driver: %s)", cfg.DBDriver)
	group := workers.StartDurableWorkers(ctx, q, cfg, svc, c.workerID)
	<-ctx.Done()

	c.Infof("Shutting down standalone durable worker...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), workerShutdownTimeout)
	defer cancel()
	group.Cancel()
	if err := group.Wait(shutdownCtx); err != nil {
		return fmt.Errorf("worker shutdown: %w", err)
	}
	return nil
}

// Usage prints command usage information for `goa4web worker`.
func (c *workerCmd) Usage() {
	fmt.Fprintf(os.Stderr, `Usage: goa4web worker [flags]

Runs standalone durable background workers against the configured database and storage.
Press Ctrl+C once for graceful shutdown; press Ctrl+C again to terminate immediately.
`)
}
