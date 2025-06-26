package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	admin "github.com/arran4/goa4web/handlers/admin"
)

// serverShutdownCmd implements "server shutdown".
type serverShutdownCmd struct {
	*serverCmd
	fs      *flag.FlagSet
	Timeout time.Duration
	args    []string
}

func parseServerShutdownCmd(parent *serverCmd, args []string) (*serverShutdownCmd, error) {
	c := &serverShutdownCmd{serverCmd: parent}
	fs := flag.NewFlagSet("shutdown", flag.ContinueOnError)
	fs.DurationVar(&c.Timeout, "timeout", 5*time.Second, "shutdown timeout")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *serverShutdownCmd) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	if err := admin.Srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
