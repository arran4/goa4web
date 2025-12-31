package dbstart

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// ensureVersionTable creates the schema_version table when missing and
// returns the current version number stored in the table.
func ensureVersionTable(ctx context.Context, db *sql.DB) (int, error) {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)"); err != nil {
		return 0, fmt.Errorf("create schema_version: %w", err)
	}
	var version int
	err := db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version)
	if err == sql.ErrNoRows {
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (?)", 1); err != nil {
			return 0, fmt.Errorf("init schema_version: %w", err)
		}
		version = 1
	} else if err != nil {
		return 0, fmt.Errorf("select schema_version: %w", err)
	}
	return version, nil
}

// Apply reads SQL migration files from the provided filesystem and executes
// each one in order, updating the schema_version table after every successful
// script. When verbose is true, progress information is printed to stdout.
func Apply(ctx context.Context, db *sql.DB, f fs.FS, verbose bool, driver string) error {
	version, err := ensureVersionTable(ctx, db)
	if err != nil {
		return err
	}
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	var nums []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		suffix := "." + driver + ".sql"
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		base := strings.TrimSuffix(name, suffix)
		n, err := strconv.Atoi(base)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)
	applied := false
	for _, n := range nums {
		if n <= version {
			continue
		}
		path := fmt.Sprintf("%04d.%s.sql", n, driver)
		if verbose {
			fmt.Printf("applying %s\n", path)
		}
		if err := executeFile(ctx, db, f, path, verbose); err != nil {
			return fmt.Errorf("execute %s: %w", path, err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE schema_version SET version = ?", n); err != nil {
			return fmt.Errorf("update schema_version: %w", err)
		}
		if verbose {
			fmt.Printf("schema version updated to %d\n", n)
		}
		applied = true
	}
	if verbose {
		if applied {
			fmt.Println("migration complete")
		} else {
			fmt.Println("database schema already up to date")
		}
	}
	return nil
}

func executeFile(ctx context.Context, db *sql.DB, f fs.FS, path string, verbose bool) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var stmt strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}
		stmt.WriteString(line)
		if strings.HasSuffix(line, ";") {
			sqlStmt := strings.TrimSuffix(stmt.String(), ";")
			if verbose {
				fmt.Printf("  %s\n", sqlStmt)
			}
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
		if verbose {
			fmt.Printf("  %s\n", s)
		}
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
