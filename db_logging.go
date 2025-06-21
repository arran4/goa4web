package main

import (
	"context"
	"database/sql/driver"
	"log"

	"github.com/segmentio/ksuid"
)

type loggingConnector struct {
	driver.Connector
}

func (lc loggingConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := lc.Connector.Connect(ctx)
	if err != nil {
		if dbLogVerbosity > 0 {
			log.Printf("DB connect error: %v", err)
		}
		return nil, err
	}
	id := ksuid.New()
	if dbLogVerbosity > 0 {
		log.Printf("db connection %s opened", id.String())
	}
	return &loggingConn{id: id, Conn: conn}, nil
}

func (lc loggingConnector) Driver() driver.Driver {
	return lc.Connector.Driver()
}

type loggingConn struct {
	id ksuid.KSUID
	driver.Conn
}

func (lc *loggingConn) Prepare(query string) (driver.Stmt, error) {
	if dbLogVerbosity >= 2 {
		log.Printf("conn %s Prepare: %s", lc.id, query)
	}
	stmt, err := lc.Conn.Prepare(query)
	if err != nil && dbLogVerbosity > 0 {
		log.Printf("conn %s Prepare error: %v", lc.id, err)
	}
	return stmt, err
}

func (lc *loggingConn) Close() error {
	if dbLogVerbosity > 0 {
		log.Printf("conn %s closed", lc.id)
	}
	return lc.Conn.Close()
}

func (lc *loggingConn) Begin() (driver.Tx, error) {
	if dbLogVerbosity >= 2 {
		log.Printf("conn %s Begin", lc.id)
	}
	tx, err := lc.Conn.Begin()
	if err != nil && dbLogVerbosity > 0 {
		log.Printf("conn %s Begin error: %v", lc.id, err)
	}
	return tx, err
}

func (lc *loggingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := lc.Conn.(driver.ConnPrepareContext); ok {
		if dbLogVerbosity >= 2 {
			log.Printf("conn %s PrepareContext: %s", lc.id, query)
		}
		stmt, err := pc.PrepareContext(ctx, query)
		if err != nil && dbLogVerbosity > 0 {
			log.Printf("conn %s PrepareContext error: %v", lc.id, err)
		}
		return stmt, err
	}
	return nil, driver.ErrSkip
}

func (lc *loggingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := lc.Conn.(driver.ExecerContext); ok {
		if dbLogVerbosity >= 2 {
			log.Printf("conn %s ExecContext: %s", lc.id, query)
		}
		res, err := ec.ExecContext(ctx, query, args)
		if err != nil && dbLogVerbosity > 0 {
			log.Printf("conn %s ExecContext error: %v", lc.id, err)
		}
		return res, err
	}
	return nil, driver.ErrSkip
}

func (lc *loggingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := lc.Conn.(driver.QueryerContext); ok {
		if dbLogVerbosity >= 2 {
			log.Printf("conn %s QueryContext: %s", lc.id, query)
		}
		rows, err := qc.QueryContext(ctx, query, args)
		if err != nil && dbLogVerbosity > 0 {
			log.Printf("conn %s QueryContext error: %v", lc.id, err)
		}
		return rows, err
	}
	return nil, driver.ErrSkip
}

func (lc *loggingConn) Ping(ctx context.Context) error {
	if p, ok := lc.Conn.(driver.Pinger); ok {
		if dbLogVerbosity >= 2 {
			log.Printf("conn %s Ping", lc.id)
		}
		err := p.Ping(ctx)
		if err != nil && dbLogVerbosity > 0 {
			log.Printf("conn %s Ping error: %v", lc.id, err)
		}
		return err
	}
	return nil
}

func (lc *loggingConn) ResetSession(ctx context.Context) error {
	if rs, ok := lc.Conn.(driver.SessionResetter); ok {
		if dbLogVerbosity >= 2 {
			log.Printf("conn %s ResetSession", lc.id)
		}
		err := rs.ResetSession(ctx)
		if err != nil && dbLogVerbosity > 0 {
			log.Printf("conn %s ResetSession error: %v", lc.id, err)
		}
		return err
	}
	return nil
}
