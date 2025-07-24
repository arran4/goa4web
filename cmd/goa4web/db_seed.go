package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
)

// dbSeedCmd implements "db seed".
type dbSeedCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
}

func parseDbSeedCmd(parent *dbCmd, args []string) (*dbSeedCmd, error) {
	c := &dbSeedCmd{dbCmd: parent}
	c.fs = newFlagSet("seed")
	c.fs.StringVar(&c.File, "file", "seed.sql", "SQL seed file")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *dbSeedCmd) Run() error {
	cfg := c.rootCmd.cfg
	conn := cfg.DBConn
	if conn == "" {
		return fmt.Errorf("connection string required")
	}
	connector, err := c.rootCmd.dbReg.Connector(cfg.DBDriver, conn)
	if err != nil {
		return err
	}
	db := sql.OpenDB(connector)
	defer db.Close()
	if err := db.Ping(); err != nil {
		return err
	}
	f, err := os.Open(c.File)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var stmt strings.Builder
	ctx := context.Background()
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}
		stmt.WriteString(line)
		if strings.HasSuffix(line, ";") {
			sqlStmt := strings.TrimSuffix(stmt.String(), ";")
			if _, err := db.ExecContext(ctx, sqlStmt); err != nil {
				return err
			}
			stmt.Reset()
		} else {
			stmt.WriteString(" ")
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if s := strings.TrimSpace(stmt.String()); s != "" {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
