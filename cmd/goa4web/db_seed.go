package main

import (
	"bufio"
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//go:embed seed.sql
var seedSql []byte

// dbSeedCmd implements "db seed".
type dbSeedCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
}

func parseDbSeedCmd(parent *dbCmd, args []string) (*dbSeedCmd, error) {
	c := &dbSeedCmd{dbCmd: parent}
	c.fs = newFlagSet("seed")
	c.fs.StringVar(&c.File, "file", "", "SQL seed file")
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
	sdb := sql.OpenDB(connector)
	defer func(sdb *sql.DB) {
		err := sdb.Close()
		if err != nil {
			log.Printf("failed to close db connection: %v", err)
		}
	}(sdb)
	if err := sdb.Ping(); err != nil {
		return err
	}
	var r io.Reader
	if c.File == "" {
		r = strings.NewReader(string(seedSql))
	} else {
		f, err := os.Open(c.File)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	scanner := bufio.NewScanner(r)
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
			if _, err := sdb.ExecContext(ctx, sqlStmt); err != nil {
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
		if _, err := sdb.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
