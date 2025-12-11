package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/arran4/goa4web/database"
)

// dbCreateCmd implements "db create".
type dbCreateCmd struct {
	*dbCmd
}

func (c *dbCreateCmd) FlagGroups() []flagGroup {
	return nil
}

var _ usageData = (*dbCreateCmd)(nil)

// Usage prints command usage.
func (c *dbCreateCmd) Usage() {
	executeUsage(c.rootCmd.fs.Output(), "db_create_usage.txt", c)
}

func parseDbCreateCmd(parent *dbCmd, args []string) (*dbCreateCmd, error) {
	c := &dbCreateCmd{dbCmd: parent}
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return c, nil
}

func (c *dbCreateCmd) Run() error {
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

	log.Println("Applying schema...")
	if err := runStatements(sdb, strings.NewReader(string(database.SchemaMySQL))); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	log.Println("Applying seed data...")
	if err := runStatements(sdb, strings.NewReader(string(database.SeedSQL))); err != nil {
		return fmt.Errorf("failed to apply seed data: %w", err)
	}

	log.Println("Database created successfully.")
	return nil
}

func runStatements(sdb *sql.DB, r io.Reader) error {
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
				return fmt.Errorf("executing statement %q: %w", sqlStmt, err)
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
			return fmt.Errorf("executing statement %q: %w", s, err)
		}
	}
	return nil
}
