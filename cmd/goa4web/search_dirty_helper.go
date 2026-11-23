package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
)

func newCLIQueries(conn *sql.DB, cfg *config.RuntimeConfig) db.Querier {
	driver := "mysql"
	if cfg != nil && cfg.DBDriver != "" {
		driver = cfg.DBDriver
	}
	return db.NewForDriver(conn, driver)
}

func withCLIJobTx(ctx context.Context, conn *sql.DB, cfg *config.RuntimeConfig, fn func(q db.Querier, svc *jobs.Service) error) error {
	q := newCLIQueries(conn, cfg)
	svc := jobs.NewConfiguredService(q)
	tx, ok := q.(db.JobTransactor)
	if !ok {
		return fmt.Errorf("durable job transactions unavailable")
	}
	return tx.WithJobTransaction(ctx, func(qtx db.Querier) error {
		return fn(qtx, svc)
	})
}
