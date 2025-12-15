package main

import (
	"database/sql"
	"fmt"
	"log"
)

func (c *rootCmd) getDB() (*sql.DB, error) {
	conn := c.cfg.DBConn
	if conn == "" {
		return nil, fmt.Errorf("connection string required")
	}
	connector, err := c.dbReg.Connector(c.cfg.DBDriver, conn)
	if err != nil {
		return nil, err
	}
	sdb := sql.OpenDB(connector)
	if err := sdb.Ping(); err != nil {
		return nil, err
	}
	return sdb, nil
}

func closeDB(sdb *sql.DB) {
	if err := sdb.Close(); err != nil {
		log.Printf("failed to close db connection: %v", err)
	}
}
